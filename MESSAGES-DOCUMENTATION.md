# Multi-lingual Messages Documentation

This document describes the multi-lingual message system implemented following Alya framework patterns.

## Overview

The application supports multi-lingual error and response messages using:
- Numeric message IDs for consistency
- Language-specific templates in `messages.json`
- Support for English (en) and Hindi (hi)
- Dynamic value substitution using placeholders

## Message Structure

### messages.json Format

```json
{
  "constants": {
    "MsgIDValidation": 101,
    "MsgIDInternalError": 102,
    // ... other constants
  },
  "messages": {
    "101": {
      "en": "{{.Field}} is {{.ErrCode}}",
      "hi": "{{.Field}} {{.ErrCode}} है"
    },
    // ... other messages
  }
}
```

### Message Components

1. **Constants Section**: Maps constant names to numeric IDs
   - Used in Go code for type safety
   - Ensures consistency across the codebase

2. **Messages Section**: Maps numeric IDs to language templates
   - Each ID has templates for supported languages
   - Templates use Go's text/template syntax

## Message IDs Used

| ID  | Constant Name | Purpose | Variables |
|-----|--------------|---------|-----------|
| 101 | MsgIDValidation | Field validation errors | Field, ErrCode |
| 102 | MsgIDInternalError | Internal server errors | vals[0] (error detail) |
| 103 | MsgIDBannedDomain | Banned email domain | Domain |
| 104 | MsgIDAlreadyExists | Duplicate email/username | Field |
| 105 | MsgIDNotFound | Resource not found | Field, vals[0] (ID) |
| 106 | MsgIDNoFieldsToUpdate | No update fields provided | None |

## Usage in Code

### Defining Constants

In `userservice.go`:
```go
const (
    MsgIDValidation       = 101
    MsgIDInternalError    = 102
    MsgIDBannedDomain     = 103
    MsgIDAlreadyExists    = 104
    MsgIDNotFound         = 105
    MsgIDNoFieldsToUpdate = 106
)
```

### Sending Error Responses

Using wscutils functions:
```go
// Validation error
wscutils.SendErrorResponse(c, wscutils.NewResponse(wscutils.ErrorStatus, nil, msgs))

// With field and error code
msgs = append(msgs, wscutils.NewErrorMessage(MsgIDValidation, 
    wscutils.ValidationErrors[v.Tag()], "Field", v.Field()))

// With custom values
msgs = append(msgs, wscutils.NewErrorMessage(MsgIDNotFound, 
    "missing", "field", "id", "vals", fmt.Sprintf("%d", req.ID)))
```

## Language Selection

The language is determined by:
1. `Accept-Language` HTTP header
2. Default to English if not specified
3. Fallback to English if translation not available

Example:
```bash
curl -X POST http://localhost:8080/users/update \
  -H "Content-Type: application/json" \
  -H "Accept-Language: hi" \
  -d '{"id": 999}'
```

## Adding New Messages

1. Add constant in Go code:
   ```go
   const MsgIDNewError = 107
   ```

2. Add to messages.json:
   ```json
   {
     "constants": {
       "MsgIDNewError": 107
     },
     "messages": {
       "107": {
         "en": "New error: {{.Detail}}",
         "hi": "नई त्रुटि: {{.Detail}}"
       }
     }
   }
   ```

3. Use in code:
   ```go
   msgs = append(msgs, wscutils.NewErrorMessage(MsgIDNewError, 
       "error_code", "Detail", "specific detail"))
   ```

## Best Practices

1. **Use Constants**: Always use named constants instead of raw numbers
2. **Consistent Variables**: Use the same variable names across similar messages
3. **Keep Messages Simple**: Avoid complex logic in templates
4. **Test All Languages**: Verify messages work in all supported languages
5. **Document Variables**: Clearly document what variables each message expects

## Testing Messages

Test different languages:
```bash
# English (default)
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{}'

# Hindi
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -H "Accept-Language: hi" \
  -d '{}'
```

## Error Code Mapping

Common validation error codes mapped by wscutils:
- `required` - Field is required
- `min` - Minimum length/value not met
- `max` - Maximum length/value exceeded
- `email` - Invalid email format
- `e164` - Invalid phone format
- `alphanum` - Non-alphanumeric characters