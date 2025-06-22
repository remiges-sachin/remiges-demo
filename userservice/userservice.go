package usersvc

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/synapsewave/remiges-demo/pg/sqlc-gen"
	"github.com/remiges-tech/alya/service"
	"github.com/remiges-tech/alya/wscutils"
	"github.com/remiges-tech/logharbour/logharbour"
)

//-----------------------------------------------------------------------------
// Constants
//-----------------------------------------------------------------------------

const (
	// Message IDs for multi-lingual support
	// These constants map to message templates in messages.json
	MsgIDValidation       = 101  // General validation errors (required, min, max, etc.)
	MsgIDInternalError    = 102  // Internal server errors
	MsgIDBannedDomain     = 103  // Email domain is banned
	MsgIDAlreadyExists    = 104  // Resource already exists (username, email)
	MsgIDNotFound         = 105  // Resource not found
	MsgIDNoFieldsToUpdate = 106  // No fields provided for update
	
	// Error codes (language-independent)
	// These are sent in the response and map to error messages in messages.json
	ErrCodeRequired      = "required"      // Field is required
	ErrCodeTooSmall      = "toosmall"      // Value too small/short (min validation)
	ErrCodeTooBig        = "toobig"        // Value too big/long (max validation)
	ErrCodeInvalidFormat = "datafmt"       // Invalid data format (email, phone, etc.)
	ErrCodeInternal      = "internal"      // Internal server error
	ErrCodeBannedDomain  = "invalid"       // Business rule violation (banned domain)
	ErrCodeAlreadyExists = "exists"        // Resource already exists
	ErrCodeNotFound      = "missing"       // Resource not found
	ErrCodeNoFields      = "missing"       // No fields provided

	// Validation constraints
	MinNameLength     = 2
	MaxNameLength     = 50
	MinUsernameLength = 3
	MaxUsernameLength = 30
	MaxEmailLength    = 100
)

//-----------------------------------------------------------------------------
// Message Templates Documentation
//-----------------------------------------------------------------------------
// See messages.json for client-side message templates in multiple languages.
// The server returns language-independent message IDs and error codes.
// Clients are responsible for formatting messages based on user language preference.

//-----------------------------------------------------------------------------
// Request Types
//-----------------------------------------------------------------------------

type CreateUserRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=50"`
	Email       string `json:"email" validate:"required,email,max=100"`
	Username    string `json:"username" validate:"required,min=3,max=30,alphanum"`
	PhoneNumber string `json:"phone_number" validate:"omitempty,e164"`
}

type UpdateUserRequest struct {
	ID          int32   `json:"id" validate:"required"`
	Name        *string `json:"name" validate:"omitempty,min=2,max=50"`
	Email       *string `json:"email" validate:"omitempty,email,max=100"`
	PhoneNumber *string `json:"phone_number" validate:"omitempty,e164"`
}

type UserResponse struct {
	ID          int32   `json:"id"`
	Name        string  `json:"name"`
	Email       string  `json:"email"`
	Username    string  `json:"username"`
	PhoneNumber *string `json:"phone_number"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

//-----------------------------------------------------------------------------
// Initialization
//-----------------------------------------------------------------------------

func init() {
	// Step 1: Set up validation tag to error code mapping
	wscutils.SetValidationTagToErrCodeMap(map[string]string{
		"required": ErrCodeRequired,
		"min":      ErrCodeTooSmall,
		"max":      ErrCodeTooBig,
		"email":    ErrCodeInvalidFormat,
		"alphanum": ErrCodeInvalidFormat,
		"e164":     ErrCodeInvalidFormat,
	})

	// Step 2: Set up validation tag to message ID mapping
	wscutils.SetValidationTagToMsgIDMap(map[string]int{
		"required": MsgIDValidation,
		"min":      MsgIDValidation,
		"max":      MsgIDValidation,
		"email":    MsgIDValidation,
		"alphanum": MsgIDValidation,
		"e164":     MsgIDValidation,
	})

	// Step 3: Set default error code and message ID
	wscutils.SetDefaultErrCode(ErrCodeInvalidFormat)
	wscutils.SetDefaultMsgID(MsgIDValidation)
}

//-----------------------------------------------------------------------------
// Request Handlers
//-----------------------------------------------------------------------------

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
		c.JSON(400, wscutils.NewResponse("error", nil, validationErrors))
		return
	}

	//-------------------------------------------------------------------------
	// Step 3: Perform business rule validations
	//-------------------------------------------------------------------------
	if isEmailDomainBanned(createUserReq.Email) {
		bannedDomainError := wscutils.BuildErrorMessage(MsgIDBannedDomain, ErrCodeBannedDomain, "email")
		c.JSON(400, wscutils.NewResponse("error", nil, []wscutils.ErrorMessage{bannedDomainError}))
		return
	}

	//-------------------------------------------------------------------------
	// Step 4: Check data dependencies
	//-------------------------------------------------------------------------
	exists, err := queries.CheckUsernameExists(c.Request.Context(), createUserReq.Username)
	if err != nil {
		logger.Error(fmt.Errorf("error checking existing user: %w", err)).LogActivity("Database error", nil)
		c.JSON(500, wscutils.NewErrorResponse(MsgIDInternalError, ErrCodeInternal))
		return
	}
	if exists {
		alreadyExistsError := wscutils.BuildErrorMessage(MsgIDAlreadyExists, ErrCodeAlreadyExists, "username")
		c.JSON(400, wscutils.NewResponse("error", nil, []wscutils.ErrorMessage{alreadyExistsError}))
		return
	}

	//-------------------------------------------------------------------------
	// Step 5: Perform core business logic
	//-------------------------------------------------------------------------
	user, err := queries.CreateUser(c.Request.Context(), sqlc.CreateUserParams{
		Name:        createUserReq.Name,
		Email:       createUserReq.Email,
		Username:    createUserReq.Username,
		PhoneNumber: sql.NullString{String: createUserReq.PhoneNumber, Valid: createUserReq.PhoneNumber != ""},
	})
	if err != nil {
		logger.Error(fmt.Errorf("error creating user: %w", err)).LogActivity("Database error", nil)
		c.JSON(500, wscutils.NewErrorResponse(MsgIDInternalError, ErrCodeInternal))
		return
	}

	//-------------------------------------------------------------------------
	// Step 6: Send response
	//-------------------------------------------------------------------------
	wscutils.SendSuccessResponse(c, wscutils.NewSuccessResponse(userToResponse(user)))
}

//-----------------------------------------------------------------------------
// Helper Functions
//-----------------------------------------------------------------------------

// Helper function to convert sqlc.User to UserResponse
func userToResponse(user sqlc.User) UserResponse {
	response := UserResponse{
		ID:       user.ID,
		Name:     user.Name,
		Email:    user.Email,
		Username: user.Username,
	}
	
	if user.PhoneNumber.Valid {
		response.PhoneNumber = &user.PhoneNumber.String
	}
	
	if user.CreatedAt.Valid {
		response.CreatedAt = user.CreatedAt.Time.Format("2006-01-02T15:04:05Z")
	}
	
	if user.UpdatedAt.Valid {
		response.UpdatedAt = user.UpdatedAt.Time.Format("2006-01-02T15:04:05Z")
	}
	
	return response
}

// Helper function to check if email domain is banned
func isEmailDomainBanned(email string) bool {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false // Malformed email will be caught by email validator
	}

	bannedDomains := []string{"banned.com", "example.com"}
	emailDomain := strings.ToLower(parts[1])
	for _, domain := range bannedDomains {
		if strings.ToLower(domain) == emailDomain {
			return true
		}
	}
	return false
}

//-----------------------------------------------------------------------------
// Update User Handler
//-----------------------------------------------------------------------------
// This handler demonstrates:
// 1. Partial updates with pointer fields
// 2. Data change logging (changelog) using LogHarbour
// 3. Activity logging for audit trails
// 4. Comprehensive validation and error handling

func HandleUpdateUserRequest(c *gin.Context, s *service.Service) {
	logger := s.LogHarbour.WithModule("UserService")
	logger.Info().LogActivity("UpdateUser request received", nil)

	// Get queries object
	queries := s.Database.(*sqlc.Queries)

	// Parse and bind request data
	var updateUserReq UpdateUserRequest
	if err := wscutils.BindJSON(c, &updateUserReq); err != nil {
		return
	}

	// Check if user exists and get current values for changelog
	currentUser, err := queries.GetUserByID(c.Request.Context(), updateUserReq.ID)
	if err != nil {
		if err == sql.ErrNoRows {
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
		updateParams.Name = sql.NullString{String: *updateUserReq.Name, Valid: true}
	}
	if updateUserReq.Email != nil {
		updateParams.Email = sql.NullString{String: *updateUserReq.Email, Valid: true}
	}
	if updateUserReq.PhoneNumber != nil {
		updateParams.PhoneNumber = sql.NullString{String: *updateUserReq.PhoneNumber, Valid: true}
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
		logger.LogDataChange(fmt.Sprintf("User %d updated", updateUserReq.ID), *changeInfo)
	}

	// Log the update activity
	logger.Info().LogActivity("User updated", map[string]any{
		"user_id": updateUserReq.ID,
		"updated_fields": map[string]any{
			"name_changed":  updateUserReq.Name != nil,
			"email_changed": updateUserReq.Email != nil,
			"phone_changed": updateUserReq.PhoneNumber != nil,
		},
	})

	// Send response
	wscutils.SendSuccessResponse(c, wscutils.NewSuccessResponse(nil))
}
