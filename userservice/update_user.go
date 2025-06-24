package usersvc

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/synapsewave/remiges-demo/pg/sqlc-gen"
	"github.com/remiges-tech/alya/service"
	"github.com/remiges-tech/alya/wscutils"
	"github.com/remiges-tech/logharbour/logharbour"
)

// HandleUpdateUserRequest demonstrates:
// 1. Partial updates with pointer fields
// 2. Data change logging (changelog) using LogHarbour
// 3. Activity logging for audit trails
// 4. Comprehensive validation and error handling
func HandleUpdateUserRequest(c *gin.Context, s *service.Service) {
	// Parse and bind request data first to get the ID
	var updateUserReq UpdateUserRequest
	if err := wscutils.BindJSON(c, &updateUserReq); err != nil {
		return
	}

	// Create logger with module and instance information
	logger := s.LogHarbour.WithModule("UserService").WithInstanceId(fmt.Sprintf("%d", updateUserReq.ID))
	logger.Info().LogActivity("UpdateUser request received", nil)

	// Get queries object
	queries := s.Database.(*sqlc.Queries)

	// Check if user exists and get current values for changelog
	currentUser, err := queries.GetUserByID(c.Request.Context(), updateUserReq.ID)
	if err != nil {
		if err == pgx.ErrNoRows {
			logger.Info().LogActivity("User not found", map[string]any{"id": updateUserReq.ID})
			notFoundError := wscutils.BuildErrorMessage(MsgIDNotFound, ErrCodeNotFound, "id", fmt.Sprintf("%d", updateUserReq.ID))
			wscutils.SendErrorResponse(c, wscutils.NewResponse("error", nil, []wscutils.ErrorMessage{notFoundError}))
			return
		}
		logger.Error(fmt.Errorf("error fetching user: %w", err)).LogActivity("Database error", nil)
		internalError := wscutils.BuildErrorMessage(MsgIDInternalError, ErrCodeInternal, "", "")
		wscutils.SendErrorResponse(c, wscutils.NewResponse("error", nil, []wscutils.ErrorMessage{internalError}))
		return
	}

	// Check if at least one field is being updated
	if updateUserReq.Name == nil && updateUserReq.Email == nil && updateUserReq.PhoneNumber == nil {
		logger.Info().LogActivity("No fields to update", nil)
		noUpdatesError := wscutils.BuildErrorMessage(MsgIDNoFieldsToUpdate, ErrCodeNoFields, "", "")
		wscutils.SendErrorResponse(c, wscutils.NewResponse("error", nil, []wscutils.ErrorMessage{noUpdatesError}))
		return
	}

	// Get validation constraints from Rigel
	minNameLength, err := s.RigelConfig.Get(c.Request.Context(), "validation.name.minLength")
	if err != nil {
		minNameLength = "2"
	}
	maxNameLength, err := s.RigelConfig.Get(c.Request.Context(), "validation.name.maxLength")
	if err != nil {
		maxNameLength = "50"
	}
	maxEmailLength, err := s.RigelConfig.Get(c.Request.Context(), "validation.email.maxLength")
	if err != nil {
		maxEmailLength = "100"
	}

	// Validate request data
	validationErrors := wscutils.WscValidate(updateUserReq, func(err validator.FieldError) []string {
		switch err.Tag() {
		case "min", "max":
			if err.Value() == nil {
				return []string{}
			}
			strVal, ok := err.Value().(*string)
			if !ok || strVal == nil {
				return []string{}
			}
			currentLen := len(*strVal)
			
			switch err.Field() {
			case "Name":
				return []string{fmt.Sprintf("%d", currentLen), minNameLength, maxNameLength}
			case "Email":
				return []string{fmt.Sprintf("%d", currentLen), "0", maxEmailLength}
			default:
				return []string{fmt.Sprintf("%d", currentLen), "0", err.Param()}
			}

		case "email":
			if err.Value() == nil {
				return []string{}
			}
			strVal, ok := err.Value().(*string)
			if !ok || strVal == nil {
				return []string{}
			}
			return []string{*strVal}

		case "e164":
			if err.Value() == nil {
				return []string{}
			}
			strVal, ok := err.Value().(*string)
			if !ok || strVal == nil {
				return []string{}
			}
			return []string{*strVal}

		default:
			return []string{}
		}
	})

	if len(validationErrors) > 0 {
		wscutils.SendErrorResponse(c, wscutils.NewResponse("error", nil, validationErrors))
		return
	}

	// Business rule validations
	if updateUserReq.Email != nil {
		// Check if email domain is banned
		if isEmailDomainBanned(*updateUserReq.Email) {
			bannedDomainError := wscutils.BuildErrorMessage(MsgIDBannedDomain, ErrCodeBannedDomain, "email")
			wscutils.SendErrorResponse(c, wscutils.NewResponse("error", nil, []wscutils.ErrorMessage{bannedDomainError}))
			return
		}

		// Check if email already exists for another user
		exists, err := queries.CheckEmailExistsForUpdate(c.Request.Context(), sqlc.CheckEmailExistsForUpdateParams{
			Email: *updateUserReq.Email,
			ID:    updateUserReq.ID,
		})
		if err != nil {
			logger.Error(fmt.Errorf("error checking email existence: %w", err)).LogActivity("Database error", nil)
			internalError := wscutils.BuildErrorMessage(MsgIDInternalError, ErrCodeInternal, "", "")
			wscutils.SendErrorResponse(c, wscutils.NewResponse("error", nil, []wscutils.ErrorMessage{internalError}))
			return
		}
		if exists {
			alreadyExistsError := wscutils.BuildErrorMessage(MsgIDAlreadyExists, ErrCodeAlreadyExists, "email")
			wscutils.SendErrorResponse(c, wscutils.NewResponse("error", nil, []wscutils.ErrorMessage{alreadyExistsError}))
			return
		}
	}

	// Prepare update parameters
	updateParams := sqlc.UpdateUserParams{
		ID: updateUserReq.ID,
	}
	
	if updateUserReq.Name != nil {
		updateParams.Name = pgtype.Text{String: *updateUserReq.Name, Valid: true}
	}
	if updateUserReq.Email != nil {
		updateParams.Email = pgtype.Text{String: *updateUserReq.Email, Valid: true}
	}
	if updateUserReq.PhoneNumber != nil {
		updateParams.PhoneNumber = pgtype.Text{String: *updateUserReq.PhoneNumber, Valid: true}
	}

	// Update user
	_, err = queries.UpdateUser(c.Request.Context(), updateParams)
	if err != nil {
		logger.Error(fmt.Errorf("error updating user: %w", err)).LogActivity("Database error", nil)
		internalError := wscutils.BuildErrorMessage(MsgIDInternalError, ErrCodeInternal, "", "")
		wscutils.SendErrorResponse(c, wscutils.NewResponse("error", nil, []wscutils.ErrorMessage{internalError}))
		return
	}

	// Create changelog for the update
	changeInfo := logharbour.NewChangeInfo("User", "Update")
	
	// Track name change
	if updateUserReq.Name != nil && currentUser.Name != *updateUserReq.Name {
		changeInfo.AddChange("name", currentUser.Name, *updateUserReq.Name)
	}
	
	// Track email change
	if updateUserReq.Email != nil && currentUser.Email != *updateUserReq.Email {
		changeInfo.AddChange("email", currentUser.Email, *updateUserReq.Email)
	}
	
	// Track phone number change
	if updateUserReq.PhoneNumber != nil {
		oldPhone := ""
		if currentUser.PhoneNumber.Valid {
			oldPhone = currentUser.PhoneNumber.String
		}
		if oldPhone != *updateUserReq.PhoneNumber {
			changeInfo.AddChange("phone_number", oldPhone, *updateUserReq.PhoneNumber)
		}
	}
	
	// Log the data change if there were actual changes
	if len(changeInfo.Changes) > 0 {
		logger.LogDataChange("User updated", *changeInfo)
	}

	// Log the update activity
	logger.Info().LogActivity("User updated", map[string]any{
		"updated_fields": map[string]any{
			"name_changed":  updateUserReq.Name != nil,
			"email_changed": updateUserReq.Email != nil,
			"phone_changed": updateUserReq.PhoneNumber != nil,
		},
	})

	// Send response
	wscutils.SendSuccessResponse(c, wscutils.NewSuccessResponse(nil))
}