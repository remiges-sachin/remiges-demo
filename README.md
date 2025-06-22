# User Service Example

This is a sample microservice built using the Remiges stack (Alya framework) that demonstrates user management functionality with:
- Dynamic configuration using Rigel
- Three-tier logging with LogHarbour (request, activity, and change logs)
- Automatic HTTP request logging middleware
- Data change tracking (changelog) for audit trails
- Comprehensive validation and error handling
- Multi-lingual message support following Alya patterns
- Centralized logging with Kafka and Elasticsearch integration
- Log visualization with Kibana

## Prerequisites

- Go 1.19 or later
- Docker and Docker Compose
- etcd (provided via Docker Compose)
- PostgreSQL (provided via Docker Compose)
- Kafka (provided via Docker Compose)
- Elasticsearch and Kibana (provided via Docker Compose)

## Project Structure

```
usersvc-example/
├── docker-compose.yaml    # Docker composition for all infrastructure
├── main.go               # Main application entry point
├── setup-config.sh       # Script to initialize Rigel configuration
├── usersvc-schema.json   # Configuration schema for Rigel
├── messages.json         # Multi-lingual message templates
├── userservice/          # User service implementation
├── consumer/             # LogHarbour Kafka consumer service
│   ├── main.go          # Consumer implementation
│   ├── Dockerfile       # Container image for consumer
│   └── go.mod           # Consumer dependencies
└── test-*.sh            # Test scripts for pipeline verification
```

## Setup and Configuration

### Quick Setup (Recommended)
```bash
# Initialize everything with one command
./scripts/init.sh
```

### Manual Setup
1. Start the required services:
   ```bash
   docker compose up -d
   ```

2. **IMPORTANT: Run database migrations**:
   ```bash
   # Install tern
   go install github.com/jackc/tern/v2@latest
   
   # Run migrations
   cd pg/migrations
   tern migrate
   cd ../..
   ```

3. Run the setup script to initialize configuration:
   ```bash
   ./setup-config.sh
   ```
   This script:
   - Waits for etcd to be ready
   - Loads the configuration schema
   - Sets up database configuration
   - Configures validation rules

4. Build and run the service:
   ```bash
   go run .
   ```

## Configuration Details

The service uses Rigel for dynamic configuration management. Key configuration parameters include:

### Database Configuration
- Host: localhost
- Port: 5432
- User: alyatest
- Password: alyatest
- Database: alyatest

## Features

### User Management
- **Create User** (POST /users) - Create new users with validation
- **Update User** (POST /users/update) - Partial updates with field-level change tracking

### Logging System
Three-tier logging system using LogHarbour with Kafka and Elasticsearch:
- **Request Logging**: Automatic HTTP request/response logging via Alya middleware
- **Activity Logging**: Business operation tracking
- **Change Logging**: Field-level data modification tracking
- **Centralized Logging**: Kafka as message broker, Elasticsearch for storage
- **Log Visualization**: Kibana for searching and analyzing logs
- See [LOGGING-SETUP.md](LOGGING-SETUP.md) for complete details
- See [CHANGELOG-FEATURE.md](CHANGELOG-FEATURE.md) for data change tracking
- See [KAFKA-ELASTICSEARCH-SETUP.md](KAFKA-ELASTICSEARCH-SETUP.md) for pipeline setup

### Validation
- Comprehensive input validation using Alya's validation framework
- Custom error messages with field-specific details
- Support for multiple validation errors in a single response

### Multi-lingual Support
- Message templates in English and Hindi
- Numeric message IDs for consistent error codes
- Extensible message system via messages.json

## Development

### Using Development Scripts
```bash
# Start application with monitoring
./scripts/dev.sh start

# View logs
./scripts/dev.sh logs

# Monitor Kafka messages
./scripts/dev.sh kafka

# Check service status
./scripts/dev.sh status

# See all available commands
./scripts/dev.sh
```

### Manual Development
1. Make sure etcd is running before starting the service
2. The setup script must be run at least once to initialize configuration
3. Any configuration changes can be made using the `rigelctl` command-line tool

### Cleanup
```bash
# Clean up everything (with prompts)
./scripts/cleanup.sh
```

## Troubleshooting

1. If the service fails to start, ensure:
   - etcd is running and accessible
   - PostgreSQL is running
   - Configuration is properly set in etcd
   - All required environment variables are set

2. To check configuration values:
   ```bash
   rigelctl --app alya --module usersvc --version 1 --config dev config get <key>
   ```

## Documentation

- [API Documentation](API-DOCUMENTATION.md) - Detailed API endpoints and examples
- [Setup Guide](SETUP.md) - Complete setup instructions with troubleshooting
- [Logging Setup](LOGGING-SETUP.md) - LogHarbour configuration details
- [Changelog Feature](CHANGELOG-FEATURE.md) - Data change tracking implementation
- [Kafka-Elasticsearch Setup](KAFKA-ELASTICSEARCH-SETUP.md) - Centralized logging pipeline
- [Messages Documentation](MESSAGES-DOCUMENTATION.md) - Multi-lingual message system
