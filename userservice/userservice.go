package usersvc

// This package implements a user service with handlers split across multiple files:
// - common.go: Shared constants, types, and helper functions
// - create_user.go: Handler for creating new users
// - get_user.go: Handler for retrieving users by ID
// - update_user.go: Handler for updating existing users
//
// The handlers demonstrate:
// - Alya framework patterns for request/response handling
// - Rigel configuration for validation constraints
// - LogHarbour for comprehensive logging (activity logs, data change logs)
// - SQLC for type-safe database operations
// - Multi-lingual error handling with message IDs
//
// All handlers follow a consistent pattern:
// 1. Parse and bind request data
// 2. Validate using both struct tags and custom validation
// 3. Apply business rules
// 4. Check data dependencies
// 5. Perform core business logic
// 6. Send appropriate response