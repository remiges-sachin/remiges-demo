# Multi-Lingual Message Support

This document explains how the user service implements Alya's multi-lingual message support pattern.

## Overview

The service returns language-independent error responses containing:
- **Message ID** (msgid): Numeric identifier for the message type
- **Error Code** (errcode): Language-independent error classification
- **Field**: The field that caused the error
- **Values** (vals): Dynamic values to be inserted into message templates

## Message IDs and Error Codes

### Message IDs Used in Responses
```go
const (
    MsgIDValidation       = 101  // General validation errors
    MsgIDInternalError    = 102  // Internal server errors
    MsgIDBannedDomain     = 103  // Email domain is banned
    MsgIDAlreadyExists    = 104  // Resource already exists
    MsgIDNotFound         = 105  // Resource not found
    MsgIDNoFieldsToUpdate = 106  // No fields provided for update
)
```

### Message Template Structure
- **101**: Special case - combines field name with error code template
- **102-106**: Direct message templates with placeholders

### Error Codes (Alya Standard)
```go
const (
    ErrCodeRequired      = "required"   // Field is required
    ErrCodeTooSmall      = "toosmall"   // Value is too small/short
    ErrCodeTooBig        = "toobig"     // Value is too big/long
    ErrCodeInvalidFormat = "datafmt"    // Invalid data format
    ErrCodeInternal      = "internal"   // Internal error
    ErrCodeBannedDomain  = "invalid"    // Invalid value (business rule)
    ErrCodeAlreadyExists = "exists"     // Resource already exists
    ErrCodeNotFound      = "missing"    // Resource not found
    ErrCodeNoFields      = "missing"    // Missing required data
)
```

## Server Response Format

### Example Error Response
```json
{
  "status": "error",
  "data": null,
  "messages": [
    {
      "msgid": 101,
      "errcode": "toosmall",
      "field": "name",
      "vals": ["1", "2", "50"]
    },
    {
      "msgid": 101,
      "errcode": "datafmt",
      "field": "email",
      "vals": ["not-an-email"]
    }
  ]
}
```

## Client-Side Implementation

The client must:

1. **Maintain Message Templates** (see messages.json)
```javascript
// Message templates by ID
const messageTemplates = {
  "101": { "en": "@<field>@ @<errcode>@", "hi": "@<field>@ @<errcode>@" },
  "102": { "en": "An internal error occurred. Please try again later", "hi": "एक आंतरिक त्रुटि हुई। कृपया बाद में पुनः प्रयास करें" },
  "103": { "en": "Email domain is not allowed", "hi": "ईमेल डोमेन की अनुमति नहीं है" },
  "104": { "en": "@<field>@ already exists", "hi": "@<field>@ पहले से मौजूद है" },
  "105": { "en": "@<field>@ not found", "hi": "@<field>@ नहीं मिला" },
  "106": { "en": "No fields provided for update", "hi": "अपडेट के लिए कोई फ़ील्ड प्रदान नहीं की गई" }
};

// Error code templates for msgid 101
const errcodeTemplates = {
  "required": { "en": "is required", "hi": "आवश्यक है" },
  "toosmall": { "en": "must be at least @<vals[1]>@ characters (current: @<vals[0]>@)", "hi": "कम से कम @<vals[1]>@ अक्षर होना चाहिए (वर्तमान: @<vals[0]>@)" }
};
```

2. **Replace Placeholders**
```javascript
function formatMessage(template, field, vals, fieldNames, lang) {
  let message = template;
  
  // Replace field placeholder with localized field name
  const localizedField = fieldNames[field][lang] || field;
  message = message.replace('@<field>@', localizedField);
  
  // Replace value placeholders
  vals.forEach((val, index) => {
    message = message.replace(`@<vals[${index}]>@`, val);
  });
  
  return message;
}
```

3. **Display Localized Messages**
```javascript
function displayError(error, lang) {
  let template = messageTemplates[error.msgid][lang];
  
  // Special handling for msgid 101 (validation errors)
  if (error.msgid === 101) {
    const errcodeTemplate = errcodeTemplates[error.errcode][lang];
    template = template.replace('@<errcode>@', errcodeTemplate);
  }
  
  return formatMessage(template, error.field, error.vals, fieldNames, lang);
}

// Example outputs:
// msgid 101: "Name is required" / "नाम आवश्यक है"
// msgid 104: "Username already exists" / "उपयोगकर्ता नाम पहले से मौजूद है"
```

## Validation Values Array Pattern

Different error types use the vals array differently:

### Length Validation (min/max)
- `vals[0]`: Current length
- `vals[1]`: Minimum allowed length
- `vals[2]`: Maximum allowed length

### Format Validation (email, phone)
- `vals[0]`: The invalid value provided

### Existence Checks
- No vals needed (field name is sufficient)

## Implementation in Handlers

### Constants to Response Mapping

When building error responses, the Go constants are used:

```go
// Example 1: Banned domain error
wscutils.BuildErrorMessage(MsgIDBannedDomain, ErrCodeBannedDomain, "email")
// Produces: { "msgid": 103, "errcode": "invalid", "field": "email" }

// Example 2: Already exists error  
wscutils.BuildErrorMessage(MsgIDAlreadyExists, ErrCodeAlreadyExists, "username")
// Produces: { "msgid": 104, "errcode": "exists", "field": "username" }

// Example 3: Validation error (handled by framework)
// When validation fails, it uses MsgIDValidation (101) with appropriate errcode
// Min validation produces: { "msgid": 101, "errcode": "toosmall", "field": "name", "vals": ["1", "2", "50"] }
```

### Setting up Validation
```go
validationErrors := wscutils.WscValidate(req, func(err validator.FieldError) []string {
    switch err.Tag() {
    case "min", "max":
        currentLen := len(err.Value().(string))
        return []string{
            fmt.Sprintf("%d", currentLen),  // Current length
            minLength,                       // Min length
            maxLength,                       // Max length
        }
    case "email":
        return []string{err.Value().(string)}  // Invalid email
    default:
        return []string{}
    }
})
```

### Building Error Messages
```go
// Simple error (no vals)
bannedError := wscutils.BuildErrorMessage(MsgIDBannedDomain, ErrCodeBannedDomain, "email")

// Error with values
notFoundError := wscutils.BuildErrorMessage(MsgIDNotFound, ErrCodeNotFound, "id", 
    fmt.Sprintf("%d", userID))
```

## Benefits

1. **Language Independence**: Server doesn't need to know about languages
2. **Maintainability**: Add new languages without changing server code
3. **Consistency**: Standard error format across all endpoints
4. **Flexibility**: Clients control message formatting and presentation
5. **Performance**: No server-side translation overhead

## Testing

To test multi-lingual support:

1. Make a request that triggers validation errors
2. Verify the response contains numeric msgid and errcode
3. Implement client-side formatting for your language
4. Test placeholder replacement with actual values

## Adding New Messages

1. Add constant in userservice.go
2. Update messages.json with templates for all languages
3. Use wscutils.BuildErrorMessage with appropriate msgid and errcode
4. Document the vals array usage for the new message type