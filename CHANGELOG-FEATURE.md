# Data Change Logging (Changelog) Feature

This document explains how the changelog feature is implemented in the user service using LogHarbour.

## Overview

The update endpoint demonstrates LogHarbour's data change logging capability, which automatically tracks:
- What was changed (field names)
- Old values vs new values
- When the change occurred
- Who made the change (when authentication is implemented)

## Implementation Details

### 1. Import LogHarbour
```go
import "github.com/remiges-tech/logharbour/logharbour"
```

### 2. Capture Current State
Before updating, we fetch the current user data to compare with new values:
```go
currentUser, err := queries.GetUserByID(c.Request.Context(), updateUserReq.ID)
```

### 3. Track Changes
After a successful update, we create a changelog entry:
```go
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
```

### 4. Log the Changes
Only log if there were actual changes:
```go
if len(changeInfo.Changes) > 0 {
    logger.LogDataChange(fmt.Sprintf("User %d updated", updateUserReq.ID), *changeInfo)
}
```

## Log Output Example

When a user is updated, LogHarbour generates a structured log entry like:

```json
{
    "id": "unique-log-id",
    "app": "UserService",
    "type": "C",
    "pri": "Info",
    "when": "2024-06-22T18:30:00Z",
    "msg": "User 1 updated",
    "data": {
        "change_data": {
            "entity": "User",
            "op": "Update",
            "changes": [
                {
                    "field": "name",
                    "old_value": "Test User",
                    "new_value": "Updated User"
                },
                {
                    "field": "email",
                    "old_value": "test@valid.com",
                    "new_value": "updated@valid.com"
                }
            ]
        }
    }
}
```

## Benefits

1. **Audit Trail**: Complete history of all data modifications
2. **Debugging**: Easy to trace when and how data changed
3. **Compliance**: Meets regulatory requirements for data tracking
4. **Analytics**: Understand user behavior and data patterns

## Querying Change Logs

LogHarbour stores these logs in Elasticsearch, allowing powerful queries:

```go
// Example: Get all changes for a specific user
// This would be implemented in a separate endpoint
changes, err := logharbour.GetChanges(queryToken, esClient, logharbour.GetLogsParam{
    App: "UserService",
    Class: strPtr("User"),
    Instance: strPtr("1"), // User ID
    Days: intPtr(30),
})
```

## Best Practices

1. **Always Compare Values**: Only log actual changes, not unchanged fields
2. **Handle Null Values**: Properly handle SQL null types when comparing
3. **Meaningful Messages**: Use clear, descriptive messages for the change
4. **Entity Consistency**: Use consistent entity names across the application
5. **Operation Types**: Use standard operation names (Create, Update, Delete)

## Future Enhancements

1. **Authentication Integration**: Once auth is added, automatically capture who made changes
2. **Batch Updates**: Track bulk operations efficiently
3. **Change Reasons**: Add support for capturing why changes were made
4. **Change Approval**: Track approval workflows for sensitive changes