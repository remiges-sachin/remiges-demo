# Setup Guide

This guide walks you through setting up the User Service with all its components including Kafka, Elasticsearch, and Kibana for centralized logging.

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

# 2. Create database tables
PGPASSWORD=alyatest psql -h localhost -U alyatest -d alyatest -f create_tables.sql

# 3. Initialize Rigel configuration
./setup-config.sh

# 4. Run the application
go run .
```

## Detailed Setup Instructions

### 1. Infrastructure Services

The `docker-compose.yaml` includes all required services:

#### Core Services
- **PostgreSQL**: Database for user data
  - Port: 5432
  - Credentials: alyatest/alyatest
  
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

Create the users table:

```sql
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    username VARCHAR(30) NOT NULL UNIQUE,
    phone_number VARCHAR(20),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

Or use the provided SQL file:
```bash
PGPASSWORD=alyatest psql -h localhost -U alyatest -d alyatest -f create_tables.sql
```

### 3. Configuration Management (Rigel)

The `setup-config.sh` script initializes all required configurations in etcd:

```bash
./setup-config.sh
```

This sets up:
- Database connection parameters
- Validation rules for email domains
- Service configuration

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
PGPASSWORD=alyatest psql -h localhost -U alyatest -d alyatest -c "SELECT 1"
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