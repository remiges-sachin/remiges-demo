# Quick Start Guide

If you've already set up the project and just need to run it:

## 1. Start Infrastructure
```bash
docker-compose up -d
```

## 2. Run Migrations (First Time Only)
```bash
# Install tern if not already installed
go install github.com/jackc/tern/v2@latest

# Run migrations
cd pg/migrations
tern migrate
cd ../..
```

## 3. Initialize Configuration (First Time Only)
```bash
./setup-config.sh
```

## 4. Run the Service
```bash
go run .
```

## 5. Test
```bash
# Create a user
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{
    "data": {
      "name": "John Doe",
      "email": "john@example.com",
      "username": "johndoe",
      "phone_number": "+1234567890"
    }
  }'

# Update a user (ID is now in the request body)
curl -X POST http://localhost:8080/users/update \
  -H "Content-Type: application/json" \
  -d '{
    "data": {
      "id": 1,
      "name": "John Updated",
      "email": "john.updated@example.com"
    }
  }'

# Run comprehensive tests
./test-update.sh
```

## Common Commands

### Check Services
```bash
docker-compose ps
```

### View Logs
```bash
# Application logs
go run . 2>&1 | tee app.log

# PostgreSQL logs
docker-compose logs postgres

# etcd logs
docker-compose logs etcd
```

### Database Access
```bash
# Connect to PostgreSQL
docker exec -it remiges-demo-postgres-1 psql -U alyatest -d alyatest

# Common SQL commands:
\dt              # List tables
\d users         # Describe users table
SELECT * FROM users;  # View all users
```

### Stop Everything
```bash
# Stop the application (Ctrl+C)

# Stop infrastructure
docker-compose down

# Stop and remove all data
docker-compose down -v
```

## Features Demonstrated

1. **Dynamic Configuration** - Server port, database settings, and validation rules from Rigel
2. **Three-Tier Logging System**:
   - HTTP request logging (automatic via Alya middleware)
   - Activity logs for business operations
   - Data change logs for audit trails
3. **Data Changelog** - Automatic tracking of field-level changes in updates
4. **Validation Framework** - Comprehensive input validation with detailed error messages
5. **Error Handling** - Consistent error responses using Alya patterns

## Troubleshooting

### "relation 'users' does not exist"
Run the migrations:
```bash
cd pg/migrations && tern migrate && cd ../..
```

### "etcd connection refused"
Make sure etcd is running:
```bash
docker-compose up -d etcd
```

### "configuration not found"
Run the setup script:
```bash
./setup-config.sh
```

### Port 8080 already in use
Either stop the other service or change the port in etcd:
```bash
rigelctl --etcd-endpoint localhost:2379 --app alya --module usersvc --version 1 --config dev config set server.port 8081
```