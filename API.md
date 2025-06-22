# User Service API Documentation

## Base URL
```
http://localhost:8080
```

## Logging
All requests are automatically logged with:
- HTTP method, path, status code
- Request/response sizes and duration
- Client IP and user agent
- See [LOGGING-SETUP.md](LOGGING-SETUP.md) for details

## Endpoints

### 1. Create User
Creates a new user in the system.

**Endpoint:** `POST /users`

**Request Body:**
```json
{
  "data": {
    "name": "John Doe",
    "email": "john@example.com",
    "username": "johndoe",
    "phone_number": "+1234567890"
  }
}
```

**Response (Success):**
```json
{
  "status": "success",
  "data": {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com",
    "username": "johndoe",
    "phone_number": "+1234567890",
    "created_at": "2024-06-22T16:45:00Z",
    "updated_at": "2024-06-22T16:45:00Z"
  }
}
```

**Validation Rules:**
- `name`: Required, 2-50 characters
- `email`: Required, valid email format, max 100 characters
- `username`: Required, 3-30 characters, alphanumeric only
- `phone_number`: Optional, E.164 format (e.g., +1234567890)

### 2. Update User
Updates an existing user's information. Supports partial updates.

**Endpoint:** `POST /users/update`

**Request Body:**
```json
{
  "data": {
    "id": 1,
    "name": "Updated Name",
    "email": "updated@example.com",
    "phone_number": "+9876543210"
  }
}
```

**Response (Success):**
```json
{
  "status": "success",
  "data": null
}
```

**Notes:**
- `id` field is required in the request body
- All other fields are optional
- Only provided fields will be updated
- Username cannot be updated
- Email must be unique across users

**Validation Rules:**
- `id`: Required, must be a valid user ID
- `name`: Optional, 2-50 characters
- `email`: Optional, valid email format, max 100 characters
- `phone_number`: Optional, E.164 format

## Error Responses

All errors follow the standard Alya error format:

```json
{
  "status": "error",
  "data": null,
  "errors": [
    {
      "msgid": 101,
      "ecode": "required",
      "field": "name",
      "vals": []
    }
  ]
}
```

### Common Error Codes
- `required`: Required field is missing
- `toosmall`: Value is too small/short
- `toobig`: Value is too big/long
- `datafmt`: Invalid data format
- `invalid`: Invalid value (business rules)
- `exists`: Resource already exists
- `missing`: Resource not found or missing data
- `internal`: Internal server error

### Message IDs
- `101`: Validation errors (required, min/max length, format)
- `102`: Internal server errors
- `103`: Banned email domain
- `104`: Resource already exists
- `105`: Resource not found
- `106`: No fields provided for update

### Multi-Lingual Support
The server returns language-independent error responses. Clients must implement message templates for each supported language. See [MULTI-LINGUAL-MESSAGES.md](MULTI-LINGUAL-MESSAGES.md) for implementation details.

### Banned Email Domains
The following email domains are not allowed:
- banned.com
- example.com

## Testing

Use the provided test script to run comprehensive tests:
```bash
./test-update.sh
```

Or test manually with curl commands shown in the examples above.