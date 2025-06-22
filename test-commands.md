# User Service Test Commands

## Important Notes

- The following email domains are banned and will be rejected: `banned.com`, `example.com`
- Use domains like `validmail.com`, `gmail.com`, or any other domain except the banned ones

## Prerequisites
1. Make sure the server is running: `go run main.go`
2. Make sure all required services are up (PostgreSQL, etcd, Kafka, etc.)
3. Install `jq` for pretty JSON output: `sudo apt install jq` (optional)

## Test Commands

### 1. Create User
```bash
curl -X POST http://localhost:8080/user_create \
  -H "Content-Type: application/json" \
  -d '{
    "data": {
      "name": "John Doe",
      "email": "john.doe@validmail.com",
      "username": "johndoe",
      "phone_number": "+1234567890"
    }
  }' | jq
```

### 2. Get User
```bash
# Replace 1 with the actual user ID returned from create
curl -X POST http://localhost:8080/user_get \
  -H "Content-Type: application/json" \
  -d '{"data": {"id": 1}}' | jq
```

### 3. Update User
```bash
# Replace 1 with the actual user ID
curl -X POST http://localhost:8080/user_update \
  -H "Content-Type: application/json" \
  -d '{
    "data": {
      "id": 1,
      "name": "John Updated",
      "email": "john.updated@validmail.com"
    }
  }' | jq
```

### 4. Test Validation Errors

#### Missing Required Fields
```bash
curl -X POST http://localhost:8080/user_create \
  -H "Content-Type: application/json" \
  -d '{
    "data": {
      "email": "test@example.com"
    }
  }' | jq
```

#### Invalid Email Format
```bash
curl -X POST http://localhost:8080/user_create \
  -H "Content-Type: application/json" \
  -d '{
    "data": {
      "name": "Test User",
      "email": "invalid-email",
      "username": "testuser"
    }
  }' | jq
```

#### Username Too Short
```bash
curl -X POST http://localhost:8080/user_create \
  -H "Content-Type: application/json" \
  -d '{
    "data": {
      "name": "Test User",
      "email": "test@example.com",
      "username": "ab"
    }
  }' | jq
```

#### Banned Email Domain
```bash
# Both banned.com and example.com are banned domains
curl -X POST http://localhost:8080/user_create \
  -H "Content-Type: application/json" \
  -d '{
    "data": {
      "name": "Banned User",
      "email": "user@banned.com",
      "username": "banneduser"
    }
  }' | jq
```

### 5. Test Error Cases

#### Get Non-existent User
```bash
curl -X POST http://localhost:8080/user_get \
  -H "Content-Type: application/json" \
  -d '{"data": {"id": 99999}}' | jq
```

#### Update Non-existent User
```bash
curl -X POST http://localhost:8080/user_update \
  -H "Content-Type: application/json" \
  -d '{
    "data": {
      "id": 99999,
      "name": "Does Not Exist"
    }
  }' | jq
```

#### Update with No Fields
```bash
curl -X POST http://localhost:8080/user_update \
  -H "Content-Type: application/json" \
  -d '{"data": {"id": 1}}' | jq
```

## Running the Test Script

For automated testing, run:
```bash
./test-user-service.sh
```

## Expected Response Formats

### Success Response
```json
{
  "status": "success",
  "data": {
    "id": 1,
    "name": "John Doe",
    "email": "john.doe@example.com",
    "username": "johndoe",
    "phone_number": "+1234567890",
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  }
}
```

### Error Response
```json
{
  "status": "error",
  "error": [
    {
      "message_id": 101,
      "error_code": "required",
      "field": "name",
      "params": []
    }
  ]
}
```

## Monitoring Logs

While testing, you can monitor:
1. Application logs in the terminal where the server is running
2. Kafka logs if LogHarbour is configured
3. PostgreSQL query logs
4. etcd logs for configuration access