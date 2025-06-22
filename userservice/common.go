package usersvc

import (
	"strings"

	"github.com/remiges-tech/alya/wscutils"
	"github.com/synapsewave/remiges-demo/pg/sqlc-gen"
)

//-----------------------------------------------------------------------------
// Constants
//-----------------------------------------------------------------------------

const (
	// Message IDs for multi-lingual support
	// These constants map to message templates in messages.json
	MsgIDValidation       = 101 // General validation errors (required, min, max, etc.)
	MsgIDInternalError    = 102 // Internal server errors
	MsgIDBannedDomain     = 103 // Email domain is banned
	MsgIDAlreadyExists    = 104 // Resource already exists (username, email)
	MsgIDNotFound         = 105 // Resource not found
	MsgIDNoFieldsToUpdate = 106 // No fields provided for update

	// Error codes
	// These are sent in the response and for machines to understand the error
	ErrCodeRequired      = "required" // Field is required
	ErrCodeTooSmall      = "toosmall" // Value too small/short (min validation)
	ErrCodeTooBig        = "toobig"   // Value too big/long (max validation)
	ErrCodeInvalidFormat = "datafmt"  // Invalid data format (email, phone, etc.)
	ErrCodeInternal      = "internal" // Internal server error
	ErrCodeBannedDomain  = "invalid"  // Business rule violation (banned domain)
	ErrCodeAlreadyExists = "exists"   // Resource already exists
	ErrCodeNotFound      = "missing"  // Resource not found
	ErrCodeNoFields      = "missing"  // No fields provided

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

type GetUserRequest struct {
	ID int32 `json:"id" validate:"required"`
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
