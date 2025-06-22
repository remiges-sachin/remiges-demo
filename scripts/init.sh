#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script configuration
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
COMPOSE_PROJECT_NAME="remiges-demo"

echo -e "${BLUE}=== Remiges Demo User Service Initialization ===${NC}\n"

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to wait for service
wait_for_service() {
    local service=$1
    local check_command=$2
    local max_attempts=30
    local attempt=0
    
    echo -n "Waiting for $service to be ready..."
    while [ $attempt -lt $max_attempts ]; do
        if eval $check_command >/dev/null 2>&1; then
            echo -e " ${GREEN}✓${NC}"
            return 0
        fi
        echo -n "."
        sleep 2
        ((attempt++))
    done
    echo -e " ${RED}✗ Timeout${NC}"
    return 1
}

# Change to project root
cd "$PROJECT_ROOT"

# Step 1: Check prerequisites
echo -e "${YELLOW}Step 1: Checking prerequisites${NC}"

if ! command_exists docker; then
    echo -e "${RED}Error: Docker is not installed${NC}"
    exit 1
fi

if ! command_exists docker-compose; then
    echo -e "${RED}Error: Docker Compose is not installed${NC}"
    exit 1
fi

if ! command_exists go; then
    echo -e "${RED}Error: Go is not installed${NC}"
    exit 1
fi

echo -e "${GREEN}✓ All prerequisites installed${NC}\n"

# Step 2: Stop any existing services
echo -e "${YELLOW}Step 2: Cleaning up existing services${NC}"
cd "$PROJECT_ROOT"
docker-compose down >/dev/null 2>&1
echo -e "${GREEN}✓ Cleanup complete${NC}\n"

# Step 3: Start infrastructure services
echo -e "${YELLOW}Step 3: Starting infrastructure services${NC}"
docker-compose up -d

# Wait for services to be ready
echo -e "\n${YELLOW}Step 4: Waiting for services to be ready${NC}"
wait_for_service "PostgreSQL" "docker exec demo-postgres pg_isready -U remiges"
wait_for_service "etcd" "curl -s http://localhost:2379/version"
wait_for_service "Kafka" "docker exec demo-kafka kafka-broker-api-versions --bootstrap-server localhost:9092"
wait_for_service "Elasticsearch" "curl -s http://localhost:9200/_cat/health"

# Step 5: Run database migrations
echo -e "\n${YELLOW}Step 5: Setting up database${NC}"

# Check if tern is installed
if ! command_exists tern; then
    echo -e "${YELLOW}Installing tern for database migrations...${NC}"
    go install github.com/jackc/tern/v2@latest
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Tern installed${NC}"
    else
        echo -e "${RED}✗ Failed to install tern${NC}"
        exit 1
    fi
fi

# Check if migrations directory exists
if [ -d "pg/migrations" ]; then
    cd pg/migrations
    
    # Run migrations
    echo "Running database migrations..."
    tern migrate
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Database migrations completed${NC}"
    else
        echo -e "${RED}✗ Database migration failed${NC}"
        cd ../..
        exit 1
    fi
    cd ../..
else
    echo -e "${RED}✗ Migrations directory not found at pg/migrations${NC}"
    echo -e "${YELLOW}Please ensure you have the complete project structure${NC}"
    exit 1
fi

# Step 6: Initialize Rigel configuration
echo -e "\n${YELLOW}Step 6: Initializing Rigel configuration${NC}"
if [ -f "$SCRIPT_DIR/setup-config.sh" ]; then
    "$SCRIPT_DIR/setup-config.sh" >/dev/null 2>&1
    echo -e "${GREEN}✓ Rigel configuration initialized${NC}"
else
    echo -e "${RED}✗ setup-config.sh not found${NC}"
fi

# Step 7: Install Go dependencies
echo -e "\n${YELLOW}Step 7: Installing Go dependencies${NC}"
go mod download
echo -e "${GREEN}✓ Dependencies installed${NC}"

# Step 8: Note about LogHarbour consumer
echo -e "\n${YELLOW}Step 8: LogHarbour Consumer${NC}"
echo -e "${YELLOW}Note: LogHarbour consumer is not included in the current docker-compose setup${NC}"
echo -e "${YELLOW}You may need to run it separately if required for your use case${NC}"

# Step 9: Kafka UI information
echo -e "\n${YELLOW}Step 9: Monitoring Tools${NC}"
echo "You can access Kafka UI at http://localhost:8090"
echo "This provides a web interface to monitor Kafka topics and messages"

# Summary
echo -e "\n${GREEN}=== Initialization Complete ===${NC}"
echo -e "\nServices running:"
echo "• PostgreSQL: localhost:5432 (user: remiges, db: userdb)"
echo "• etcd: localhost:2379"
echo "• Kafka: localhost:9092"
echo "• Zookeeper: localhost:2181"
echo "• Elasticsearch: localhost:9200"
echo "• Kafka UI: localhost:8090"

echo -e "\n${YELLOW}To start the application:${NC}"
echo "  go run ."

echo -e "\n${YELLOW}To run tests:${NC}"
echo "  $SCRIPT_DIR/test-complete-pipeline.sh"

echo -e "\n${YELLOW}To verify setup:${NC}"
echo "  $SCRIPT_DIR/verify-pipeline.sh"

echo -e "\n${GREEN}Happy coding! 🚀${NC}"