package usersvc

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/synapsewave/remiges-demo/pg/sqlc-gen"
	"github.com/remiges-tech/alya/service"
	"github.com/remiges-tech/alya/wscutils"
)

func HandleCreateUserRequest(c *gin.Context, s *service.Service) {
	logger := s.LogHarbour.WithModule("UserService")
	logger.Info().LogActivity("CreateUser request received", nil)

	// Get queries object once at the start
	queries := s.Database.(*sqlc.Queries)

	// Get validation constraints from Rigel
	minNameLength, err := s.RigelConfig.Get(c.Request.Context(), "validation.name.minLength")
	if err != nil {
		minNameLength = "2" // Default value
	}
	maxNameLength, err := s.RigelConfig.Get(c.Request.Context(), "validation.name.maxLength")
	if err != nil {
		maxNameLength = "50" // Default value
	}
	minUsernameLength, err := s.RigelConfig.Get(c.Request.Context(), "validation.username.minLength")
	if err != nil {
		minUsernameLength = "3" // Default value
	}
	maxUsernameLength, err := s.RigelConfig.Get(c.Request.Context(), "validation.username.maxLength")
	if err != nil {
		maxUsernameLength = "30" // Default value
	}
	maxEmailLength, err := s.RigelConfig.Get(c.Request.Context(), "validation.email.maxLength")
	if err != nil {
		maxEmailLength = "100" // Default value
	}

	//-------------------------------------------------------------------------
	// Step 1: Parse and bind request data
	//-------------------------------------------------------------------------
	var createUserReq CreateUserRequest
	if err := wscutils.BindJSON(c, &createUserReq); err != nil {
		return
	}
	logger.Info().LogActivity("CreateUser request parsed", map[string]any{"username": createUserReq.Name})

	//-------------------------------------------------------------------------
	// Step 2: Validate request data
	//-------------------------------------------------------------------------
	validationErrors := wscutils.WscValidate(createUserReq, func(err validator.FieldError) []string {
		switch err.Tag() {
		case "required":
			return []string{} // Field name is already in ErrorMessage.field

		case "min", "max":
			currentLen := len(err.Value().(string))
			switch err.Field() {
			case "Name":
				return []string{fmt.Sprintf("%d", currentLen), minNameLength, maxNameLength}
			case "Username":
				return []string{fmt.Sprintf("%d", currentLen), minUsernameLength, maxUsernameLength}
			case "Email":
				return []string{fmt.Sprintf("%d", currentLen), "0", maxEmailLength}
			default:
				return []string{fmt.Sprintf("%d", currentLen), "0", err.Param()}
			}

		case "email":
			return []string{err.Value().(string)}

		case "alphanum":
			return []string{err.Value().(string)}

		case "e164":
			return []string{err.Value().(string)}

		default:
			return []string{}
		}
	})

	if len(validationErrors) > 0 {
		wscutils.SendErrorResponse(c, wscutils.NewResponse("error", nil, validationErrors))
		return
	}

	//-------------------------------------------------------------------------
	// Step 3: Perform business rule validations
	//-------------------------------------------------------------------------
	if isEmailDomainBanned(createUserReq.Email) {
		bannedDomainError := wscutils.BuildErrorMessage(MsgIDBannedDomain, ErrCodeBannedDomain, "email")
		wscutils.SendErrorResponse(c, wscutils.NewResponse("error", nil, []wscutils.ErrorMessage{bannedDomainError}))
		return
	}

	//-------------------------------------------------------------------------
	// Step 4: Check data dependencies
	//-------------------------------------------------------------------------
	exists, err := queries.CheckUsernameExists(c.Request.Context(), createUserReq.Username)
	if err != nil {
		logger.Error(fmt.Errorf("error checking existing user: %w", err)).LogActivity("Database error", nil)
		internalError := wscutils.BuildErrorMessage(MsgIDInternalError, ErrCodeInternal, "", "")
		wscutils.SendErrorResponse(c, wscutils.NewResponse("error", nil, []wscutils.ErrorMessage{internalError}))
		return
	}
	if exists {
		alreadyExistsError := wscutils.BuildErrorMessage(MsgIDAlreadyExists, ErrCodeAlreadyExists, "username")
		wscutils.SendErrorResponse(c, wscutils.NewResponse("error", nil, []wscutils.ErrorMessage{alreadyExistsError}))
		return
	}

	//-------------------------------------------------------------------------
	// Step 5: Perform core business logic
	//-------------------------------------------------------------------------
	user, err := queries.CreateUser(c.Request.Context(), sqlc.CreateUserParams{
		Name:        createUserReq.Name,
		Email:       createUserReq.Email,
		Username:    createUserReq.Username,
		PhoneNumber: pgtype.Text{String: createUserReq.PhoneNumber, Valid: createUserReq.PhoneNumber != ""},
	})
	if err != nil {
		logger.Error(fmt.Errorf("error creating user: %w", err)).LogActivity("Database error", nil)
		internalError := wscutils.BuildErrorMessage(MsgIDInternalError, ErrCodeInternal, "", "")
		wscutils.SendErrorResponse(c, wscutils.NewResponse("error", nil, []wscutils.ErrorMessage{internalError}))
		return
	}

	//-------------------------------------------------------------------------
	// Step 6: Send response
	//-------------------------------------------------------------------------
	wscutils.SendSuccessResponse(c, wscutils.NewSuccessResponse(userToResponse(user)))
}