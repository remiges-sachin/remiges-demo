# Testing the User Service

## Quick Start

### 1. Start Infrastructure
```bash
# Start all required services
docker-compose up -d

# Wait for services to be ready (about 30 seconds)
docker-compose ps

# Check logs if needed
docker-compose logs -f
```

### 2. Build rigelctl (if not already built)
```bash
cd rigel/cmd/rigelctl
go build
cd ../../..
```

### 3. Initialize Configuration
```bash
# Run the setup script to configure etcd
./setup-config.sh
```

### 4. Run the Application
```bash
# In a new terminal, run the main application
go run main.go
```

### 5. Run Tests
```bash
# Option 1: Run automated test script
./test-user-service.sh

# Option 2: Run individual curl commands from test-commands.md
```

## Service Architecture

The user service demonstrates integration of all Remiges products:

1. **Alya** - Web framework for request/response handling
2. **Rigel** - Configuration management (database settings, validation rules)
3. **LogHarbour** - Comprehensive logging (activity logs, data change logs)
4. **PostgreSQL** - Data persistence via SQLC

## Endpoints

- `POST /user_create` - Create a new user
- `POST /user_get` - Get user by ID
- `POST /user_update` - Update user details

## Response Format

### Success Response
```json
{
  "status": "success",
  "data": { ... }
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

## Monitoring

1. **Application Logs** - Check terminal running `go run main.go`
2. **Kafka Messages** - View at http://localhost:8090 (Kafka UI)
3. **Elasticsearch** - Query at http://localhost:9200
4. **etcd** - Configuration values via rigelctl

## Troubleshooting

### Database Connection Issues
```bash
# Check PostgreSQL is running
docker-compose ps postgres

# Check database exists
docker exec -it demo-postgres psql -U remiges -d userdb -c "\dt"
```

### Configuration Issues
```bash
# List all configurations
./rigel/cmd/rigelctl/rigelctl list --endpoints localhost:2379 --app alya --sver 1 --env dev

# Get specific value
./rigel/cmd/rigelctl/rigelctl get --endpoints localhost:2379 --app alya --sver 1 --env dev --key database.host
```

### Kafka Issues
```bash
# Check Kafka is running
docker-compose logs kafka

# List topics
docker exec -it demo-kafka kafka-topics --list --bootstrap-server localhost:9092
```

## Cleanup

```bash
# Stop all services
docker-compose down

# Remove all data (careful!)
docker-compose down -v
```