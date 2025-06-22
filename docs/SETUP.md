# Setup Guide

This guide walks you through setting up the User Service with all its components including Kafka, Elasticsearch, and Kibana for centralized logging.

## Prerequisites

1. **Go 1.19 or later**
2. **Docker and Docker Compose**
3. **Required Go tools:**
   ```bash
   # Install tern for database migrations
   go install github.com/jackc/tern/v2@latest
   
   # Ensure Go bin is in PATH
   export PATH=$PATH:$(go env GOPATH)/bin
   ```
4. **rigelctl** should be available (comes with Rigel installation)

## Quick Start (Recommended)

```bash
# Run the initialization script that handles everything
./scripts/init.sh

# Then start the application
go run .
```

## Manual Setup

If you prefer to set up manually or the init script fails:

```bash
# 1. Start all infrastructure services
docker-compose up -d

# 2. Install required tools (if not already installed)
go install github.com/jackc/tern/v2@latest

# 3. Initialize Rigel configuration and run database migrations
./setup-config.sh

# 4. Run the application
go run .
```

Note: The setup-config.sh script now handles both configuration and database migrations automatically.

## Detailed Setup Instructions

### 1. Infrastructure Services

The `docker-compose.yaml` includes all required services:

#### Core Services
- **PostgreSQL**: Database for user data
  - Port: 5432
  - Credentials: remiges/remiges123
  - Database: userdb
  
- **Redis**: Caching (if needed)
  - Port: 6379
  
- **etcd**: Configuration storage for Rigel
  - Port: 2379

#### Logging Infrastructure
- **Zookeeper**: Kafka coordinator
  - Port: 2181
  
- **Kafka**: Message broker for logs
  - Port: 9092 (external), 29092 (internal)
  - Topic: `logharbour-logs` (auto-created)
  
- **Elasticsearch**: Log storage and search
  - Port: 9200 (API), 9300 (transport)
  - Single-node setup for development
  
- **Kibana**: Log visualization
  - Port: 5601
  - URL: http://localhost:5601
  
- **LogHarbour Consumer**: Indexes logs from Kafka to Elasticsearch
  - Runs as a container service

### 2. Database Setup

Database migrations are managed using **tern** and are automatically run by the setup script.

Migration files are located in `pg/migrations/`:
- `001_alyatest.sql` - Creates initial users table
- `002_alyatest.sql` - Adds username, phone_number, and timestamps
- `003_add_unique_constraints.sql` - Adds unique constraints

To manually run migrations:
```bash
cd pg/migrations
tern migrate
```

The tern configuration (`pg/migrations/tern.conf`) uses these credentials:
- Host: localhost
- Port: 5432
- User: remiges
- Password: remiges123
- Database: userdb

### 3. Configuration Management (Rigel)

The `setup-config.sh` script handles both configuration and database setup:

```bash
./setup-config.sh
```

This script:
1. Loads the configuration schema into Rigel
2. Sets up database connection parameters (matching docker-compose.yml)
3. Configures validation rules
4. Runs database migrations using tern

Key configurations set:
- Database: remiges/remiges123@localhost:5432/userdb
- Server port: 8080
- Validation rules for names, usernames, and emails

To manually check or update configuration:
```bash
# Get a config value
rigelctl --app alya --module usersvc --version 1 --config dev config get db.host

# Set a config value
rigelctl --app alya --module usersvc --version 1 --config dev config set db.host localhost
```

### 4. Application Dependencies

Install Go dependencies:
```bash
go mod download
```

### 5. Running the Application

Start the main service:
```bash
go run .
```

The service will:
1. Connect to PostgreSQL
2. Initialize Rigel client for configuration
3. Set up LogHarbour with Kafka writer (fallback to stdout)
4. Start HTTP server on port 8080

### 6. Verifying the Setup

Run the verification script:
```bash
./verify-pipeline.sh
```

This checks:
- All services are healthy
- Kafka is receiving messages
- Elasticsearch is indexing logs
- Consumer service is running

### 7. Setting Up Kibana

1. Access Kibana at http://localhost:5601
2. Wait for Kibana to initialize (may take a minute)
3. Go to **Stack Management** → **Index Patterns**
4. Create an index pattern:
   - Index pattern: `logharbour-*`
   - Time field: `when`
5. Go to **Discover** to view logs

## Testing the Pipeline

Use the provided test scripts:

```bash
# Complete pipeline test
./test-complete-pipeline.sh

# Simple verification
./verify-pipeline.sh
```

## Troubleshooting

### Service Won't Start
```bash
# Check if all containers are running
docker-compose ps

# Check logs
docker-compose logs [service-name]
```

### Database Connection Issues
```bash
# Test PostgreSQL connection
PGPASSWORD=remiges123 psql -h localhost -U remiges -d userdb -c "SELECT 1"

# If authentication fails, ensure setup-config.sh was run:
./setup-config.sh
```

### Kafka Issues
```bash
# List topics
docker exec alyatest-kafka kafka-topics --bootstrap-server localhost:9092 --list

# Check consumer group
docker exec alyatest-kafka kafka-consumer-groups --bootstrap-server localhost:9092 --group logharbour-consumer --describe
```

### Elasticsearch Issues
```bash
# Check health
curl -X GET "localhost:9200/_cat/health?v"

# Check indices
curl -X GET "localhost:9200/_cat/indices?v"
```

### Consumer Not Processing
```bash
# Check consumer logs
docker logs alyatest-logharbour-consumer

# Restart consumer
docker-compose restart logharbour-consumer
```

## Stopping Services

To stop all services:
```bash
docker-compose down
```

To stop and remove all data:
```bash
docker-compose down -v
```

## Development Tips

### Using Development Scripts

The project includes helpful development scripts in the `scripts/` directory:

```bash
# Start application
./scripts/dev.sh start

# Monitor logs
./scripts/dev.sh logs

# Watch Kafka messages
./scripts/dev.sh kafka

# Query Elasticsearch
./scripts/dev.sh elastic

# Check service status
./scripts/dev.sh status

# Clean up everything
./scripts/cleanup.sh
```

### Manual Commands

1. **Log Monitoring**: Watch logs in real-time
   ```bash
   docker logs -f alyatest-logharbour-consumer
   ```

2. **Kafka Messages**: Monitor Kafka messages
   ```bash
   docker exec alyatest-kafka kafka-console-consumer \
     --bootstrap-server localhost:9092 \
     --topic logharbour-logs \
     --from-beginning
   ```

3. **Elasticsearch Queries**: Query logs directly
   ```bash
   curl -X GET "localhost:9200/logharbour-a-*/_search?pretty" \
     -H 'Content-Type: application/json' \
     -d '{"query": {"match_all": {}}}'
   ```