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

echo -e "${BLUE}=== Remiges Demo Development Helper ===${NC}\n"

# Function to show usage
show_usage() {
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  start       Start the application with live reload"
    echo "  logs        Show application logs"
    echo "  test        Run the pipeline test"
    echo "  verify      Verify the setup"
    echo "  kafka       Monitor Kafka messages"
    echo "  elastic     Query Elasticsearch logs"
    echo "  consumer    Show consumer logs"
    echo "  restart     Restart all services"
    echo "  status      Show service status"
    echo ""
    exit 1
}

# Change to project root
cd "$PROJECT_ROOT"

# Handle commands
case "${1:-}" in
    start)
        echo -e "${YELLOW}Starting application...${NC}"
        # Check if air is installed for live reload
        if command -v air >/dev/null 2>&1; then
            echo "Using air for live reload..."
            air
        else
            echo "Starting without live reload (install 'air' for live reload)..."
            echo "To install air: go install github.com/cosmtrek/air@latest"
            go run .
        fi
        ;;
        
    logs)
        echo -e "${YELLOW}Showing application logs...${NC}"
        if [ -f app.log ]; then
            tail -f app.log
        else
            echo "No app.log file found. Start the application first."
        fi
        ;;
        
    test)
        echo -e "${YELLOW}Running pipeline test...${NC}"
        if [ -f test-complete-pipeline.sh ]; then
            ./test-complete-pipeline.sh
        else
            echo "Test script not found!"
        fi
        ;;
        
    verify)
        echo -e "${YELLOW}Verifying setup...${NC}"
        if [ -f verify-pipeline.sh ]; then
            ./verify-pipeline.sh
        else
            echo "Verification script not found!"
        fi
        ;;
        
    kafka)
        echo -e "${YELLOW}Monitoring Kafka messages (Ctrl+C to stop)...${NC}"
        docker exec alyatest-kafka kafka-console-consumer \
            --bootstrap-server localhost:9092 \
            --topic logharbour-logs \
            --from-beginning | jq .
        ;;
        
    elastic)
        echo -e "${YELLOW}Recent logs from Elasticsearch:${NC}"
        curl -s -X GET "localhost:9200/logharbour-*/_search?size=10&sort=when:desc" \
            -H 'Content-Type: application/json' | \
            jq '.hits.hits[]._source | {type, module, msg, when}'
        ;;
        
    consumer)
        echo -e "${YELLOW}Consumer logs (last 20 lines):${NC}"
        docker logs --tail 20 -f alyatest-logharbour-consumer
        ;;
        
    restart)
        echo -e "${YELLOW}Restarting all services...${NC}"
        docker-compose restart
        echo -e "${GREEN}✓ Services restarted${NC}"
        ;;
        
    status)
        echo -e "${YELLOW}Service Status:${NC}"
        docker-compose ps
        echo ""
        echo -e "${YELLOW}Quick Health Check:${NC}"
        
        # Check each service
        echo -n "PostgreSQL: "
        docker exec alyatest-pg pg_isready -U alyatest >/dev/null 2>&1 && echo -e "${GREEN}✓${NC}" || echo -e "${RED}✗${NC}"
        
        echo -n "Redis: "
        docker exec alyatest-redis redis-cli ping >/dev/null 2>&1 && echo -e "${GREEN}✓${NC}" || echo -e "${RED}✗${NC}"
        
        echo -n "Kafka: "
        docker exec alyatest-kafka kafka-broker-api-versions --bootstrap-server localhost:9092 >/dev/null 2>&1 && echo -e "${GREEN}✓${NC}" || echo -e "${RED}✗${NC}"
        
        echo -n "Elasticsearch: "
        curl -s http://localhost:9200/_cat/health >/dev/null 2>&1 && echo -e "${GREEN}✓${NC}" || echo -e "${RED}✗${NC}"
        
        echo -n "Kibana: "
        curl -s http://localhost:5601/api/status >/dev/null 2>&1 && echo -e "${GREEN}✓${NC}" || echo -e "${RED}✗${NC}"
        ;;
        
    *)
        show_usage
        ;;
esac