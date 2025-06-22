package usersvc

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/synapsewave/remiges-demo/pg/sqlc-gen"
	"github.com/remiges-tech/alya/service"
	"github.com/remiges-tech/alya/wscutils"
)

// HandleGetUserRequest retrieves a user by ID
// Demonstrates:
// 1. Simple GET operation with ID parameter
// 2. Error handling for not found cases
// 3. Activity logging for audit trails
func HandleGetUserRequest(c *gin.Context, s *service.Service) {
	logger := s.LogHarbour.WithModule("UserService")
	logger.Info().LogActivity("GetUser request received", nil)

	// Get queries object
	queries := s.Database.(*sqlc.Queries)

	// Parse and bind request data
	var getUserReq GetUserRequest
	if err := wscutils.BindJSON(c, &getUserReq); err != nil {
		return
	}

	// Validate request data
	validationErrors := wscutils.WscValidate(getUserReq, func(err validator.FieldError) []string {
		switch err.Tag() {
		case "required":
			return []string{} // Field name is already in ErrorMessage.field
		default:
			return []string{}
		}
	})

	if len(validationErrors) > 0 {
		wscutils.SendErrorResponse(c, wscutils.NewResponse("error", nil, validationErrors))
		return
	}

	// Get user from database
	user, err := queries.GetUserByID(c.Request.Context(), getUserReq.ID)
	if err != nil {
		if err == pgx.ErrNoRows {
			logger.Info().LogActivity("User not found", map[string]any{"id": getUserReq.ID})
			notFoundError := wscutils.BuildErrorMessage(MsgIDNotFound, ErrCodeNotFound, "id", fmt.Sprintf("%d", getUserReq.ID))
			wscutils.SendErrorResponse(c, wscutils.NewResponse("error", nil, []wscutils.ErrorMessage{notFoundError}))
			return
		}
		logger.Error(fmt.Errorf("error fetching user: %w", err)).LogActivity("Database error", nil)
		internalError := wscutils.BuildErrorMessage(MsgIDInternalError, ErrCodeInternal, "", "")
		wscutils.SendErrorResponse(c, wscutils.NewResponse("error", nil, []wscutils.ErrorMessage{internalError}))
		return
	}

	// Log the successful retrieval
	logger.Info().LogActivity("User retrieved", map[string]any{
		"user_id": user.ID,
		"username": user.Username,
	})

	// Send response
	wscutils.SendSuccessResponse(c, wscutils.NewSuccessResponse(userToResponse(user)))
}