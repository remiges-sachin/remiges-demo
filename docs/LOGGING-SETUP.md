# Logging Setup

This document explains the three-tier logging approach used in the Remiges demo service.

## 1. Request Logging (HTTP Middleware)

Alya's `LogRequest` middleware automatically logs every HTTP request and response.

### Setup:
```go
// Create adapter
logAdapter := router.NewLogHarbourAdapter(logger)

// Add middleware
r := gin.New()
r.Use(router.LogRequest(logAdapter))
```

### Captured Information:
- **Request**: Method, path, query params, headers, client IP, request size
- **Response**: Status code, response size, duration
- **Tracing**: X-Trace-ID and X-Span-ID headers for distributed tracing

### Example Output:
```
{"id":"abc123","app":"UserService","type":"D","pri":"Info","when":"2024-06-22T18:30:00Z","msg":"HTTP Request","data":{"method":"POST","path":"/user_create","status":200,"duration":23.45}}
```

## 2. Activity Logging (Business Logic)

Activity logs track business operations and their outcomes.

### Usage:
```go
logger.Info().LogActivity("User created", map[string]any{
    "user_id": user.ID,
    "username": user.Username,
})
```

### When to Use:
- Major business operations (create, update, delete)
- Authentication events
- Authorization decisions
- External API calls
- Batch operations

## 3. Data Change Logging (Audit Trail)

Comprehensive tracking of all data modifications for audit and compliance.

### Usage:
```go
changeInfo := logharbour.NewChangeInfo("User", "Update")
changeInfo.AddChange("email", oldEmail, newEmail)
logger.LogDataChange("User profile updated", *changeInfo)
```

### Features:
- Field-level change tracking
- Old and new value comparison
- Entity and operation type recording
- Automatic timestamp and user tracking (when auth is added)

## Log Types Summary

| Type | Code | Purpose | Example |
|------|------|---------|---------|
| Debug | D | HTTP requests, technical details | Request logging |
| Activity | A | Business operations | "User created" |
| Change | C | Data modifications | Field changes |

## Benefits of Three-Tier Logging

1. **Performance Monitoring**: Request logs show response times and status codes
2. **Business Intelligence**: Activity logs reveal user behavior patterns
3. **Compliance**: Change logs provide complete audit trail
4. **Debugging**: All three tiers help trace issues from HTTP to data layer
5. **Security**: Track suspicious activities across all layers

## Viewing Logs

During development, logs are output to stdout. In production:
- Logs are sent to Elasticsearch via Kafka
- Can be queried using LogHarbour's API
- Visualized in Kibana or similar tools

### Example: Finding Slow Requests
```bash
# Look for requests with duration > 1000ms
cat app.log | jq 'select(.data.duration > 1000)'
```

### Example: Tracking User Changes
```bash
# Find all changes for a specific user
cat app.log | jq 'select(.type == "C" and .data.change_data.entity == "User")'
```

## Best Practices

1. **Don't Log Sensitive Data**: Avoid logging passwords, tokens, or PII
2. **Use Appropriate Log Levels**: Info for normal operations, Error for failures
3. **Include Context**: Always add relevant IDs and metadata
4. **Be Consistent**: Use standard field names across the application
5. **Log Early**: Capture request context at the beginning of operations