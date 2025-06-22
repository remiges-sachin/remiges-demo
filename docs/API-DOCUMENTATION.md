# User Service API Documentation

## Base URL
```
http://localhost:8080
```

## Authentication
Currently, no authentication is required (for demo purposes).

## Endpoints

### 1. Create User
Creates a new user in the system.

**Endpoint:** `POST /user_create`

**Request Body:**
```json
{
  "name": "string",         // Required, 2-50 characters
  "email": "string",        // Required, valid email, max 100 characters
  "username": "string",     // Required, 3-30 characters, alphanumeric
  "phone_number": "string"  // Optional, E.164 format (e.g., +1234567890)
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
    "created_at": "2024-06-22T10:00:00Z",
    "updated_at": "2024-06-22T10:00:00Z"
  },
  "messages": []
}
```

**Response (Validation Error):**
```json
{
  "status": "error",
  "data": null,
  "messages": [
    {
      "msgid": 101,
      "errcode": "required",
      "field": "Name"
    },
    {
      "msgid": 101,
      "errcode": "email",
      "field": "Email"
    }
  ]
}
```

**Logs Generated:**
- Activity Log: User creation attempt
- Change Log: New user record (on success)
- Request Log: HTTP request details (automatic via middleware)

### 2. Update User
Updates an existing user with partial field updates.

**Endpoint:** `POST /user_update`

**Request Body:**
```json
{
  "id": 1,                  // Required, user ID
  "name": "string",         // Optional, 2-50 characters
  "email": "string",        // Optional, valid email, max 100 characters
  "phone_number": "string"  // Optional, E.164 format
}
```

**Note:** At least one field besides `id` must be provided for update.

**Response (Success):**
```json
{
  "status": "success",
  "data": {
    "id": 1,
    "name": "John Smith",
    "email": "john.smith@example.com",
    "username": "johndoe",
    "phone_number": "+1234567890",
    "created_at": "2024-06-22T10:00:00Z",
    "updated_at": "2024-06-22T10:30:00Z"
  },
  "messages": []
}
```

**Response (User Not Found):**
```json
{
  "status": "error",
  "data": null,
  "messages": [
    {
      "msgid": 105,
      "errcode": "missing",
      "field": "id",
      "vals": ["999"]
    }
  ]
}
```

**Response (No Fields to Update):**
```json
{
  "status": "error",
  "data": null,
  "messages": [
    {
      "msgid": 106,
      "errcode": "no_update"
    }
  ]
}
```

**Logs Generated:**
- Activity Log: Update attempt with details
- Change Log: Field-level changes (on success)
  - Old values vs new values for each changed field
  - User who made the change
  - Timestamp of change
- Request Log: HTTP request details (automatic via middleware)

## Error Codes

### Message IDs
- `101`: Validation error
- `102`: Internal server error
- `103`: Banned email domain
- `104`: Email/username already exists
- `105`: User not found
- `106`: No fields provided for update

### Validation Error Codes
- `required`: Field is required
- `min`: Value is too short
- `max`: Value is too long
- `email`: Invalid email format
- `e164`: Invalid phone number format
- `alphanum`: Only alphanumeric characters allowed

## Logging

All API requests generate logs that are:
1. Written to stdout (console)
2. Sent to Kafka topic `logharbour-logs`
3. Indexed in Elasticsearch by the consumer service
4. Viewable in Kibana at http://localhost:5601

### Log Types
- **Type A (Activity)**: Business operations
- **Type C (Change)**: Data modifications
- **Type D (Debug)**: HTTP request/response details

### Log Structure
```json
{
  "id": "unique-log-id",
  "app": "UserService",
  "system": "workstation",
  "module": "UserService",
  "type": "A|C|D",
  "pri": "Info|Warn|Err|Crit",
  "when": "2024-06-22T10:00:00Z",
  "who": "user@example.com",
  "remote_ip": "127.0.0.1",
  "trace_id": "request-trace-id",
  "msg": "Log message",
  "data": {
    // Additional structured data
  }
}
```

## Multi-lingual Support

Error messages support multiple languages via the `Accept-Language` header:
- `en` - English (default)
- `hi` - Hindi

Example:
```
Accept-Language: hi
```

Message templates are defined in `messages.json` with numeric IDs for consistency across languages.