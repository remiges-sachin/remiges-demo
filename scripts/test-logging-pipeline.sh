#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Testing LogHarbour Kafka-Elasticsearch Pipeline ===${NC}\n"

# Function to check service health
check_service() {
    local service=$1
    local check_command=$2
    
    echo -n "Checking $service... "
    if eval $check_command > /dev/null 2>&1; then
        echo -e "${GREEN}✓ OK${NC}"
        return 0
    else
        echo -e "${RED}✗ Failed${NC}"
        return 1
    fi
}

# Check all services
echo -e "${YELLOW}Step 1: Verifying infrastructure services${NC}"
check_service "PostgreSQL" "docker exec demo-postgres pg_isready -U remiges"
check_service "etcd" "curl -s http://localhost:2379/version"
check_service "Kafka" "docker exec demo-kafka kafka-broker-api-versions --bootstrap-server localhost:9092"
check_service "Elasticsearch" "curl -s http://localhost:9200/_cat/health"

echo -e "\n${YELLOW}Step 2: Starting the application${NC}"
# Start the application in background
echo "Starting user service..."
go run . > app.log 2>&1 &
APP_PID=$!
echo "Application PID: $APP_PID"

# Wait for application to start
sleep 5

# Check if app is running
if ps -p $APP_PID > /dev/null; then
    echo -e "${GREEN}✓ Application started successfully${NC}"
else
    echo -e "${RED}✗ Application failed to start${NC}"
    echo "Application logs:"
    cat app.log
    exit 1
fi

echo -e "\n${YELLOW}Step 3: Testing API endpoints to generate logs${NC}"

# Function to make API call and show result
test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local description=$4
    
    echo -e "\n${BLUE}Testing: $description${NC}"
    echo "Request: $method $endpoint"
    if [ ! -z "$data" ]; then
        echo "Data: $data"
    fi
    
    if [ "$method" = "GET" ]; then
        response=$(curl -s -X $method http://localhost:8080$endpoint)
    else
        response=$(curl -s -X $method http://localhost:8080$endpoint \
            -H "Content-Type: application/json" \
            -d "$data")
    fi
    
    echo "Response: $response"
}

# Test various endpoints to generate different log types
test_endpoint "POST" "/user_create" '{"data":{"name":"Test User","email":"test@validmail.com","username":"testuser","phone_number":"+1234567890"}}' "Create user (generates activity + change log)"
sleep 1

test_endpoint "POST" "/user_update" '{"data":{"id":1,"name":"Updated User"}}' "Update user (generates activity + change log)"
sleep 1

test_endpoint "POST" "/user_update" '{"data":{"id":999,"name":"Non-existent"}}' "Update non-existent user (generates error log)"
sleep 1

test_endpoint "POST" "/user_create" '{"data":{"name":"X","email":"invalid-email","username":"x"}}' "Invalid user data (generates validation error log)"
sleep 1

echo -e "\n${YELLOW}Step 4: Checking Kafka messages${NC}"
echo "Checking last 5 messages in Kafka topic..."
docker exec demo-kafka kafka-console-consumer \
    --bootstrap-server localhost:9092 \
    --topic logharbour-logs \
    --max-messages 5 \
    --from-beginning \
    --timeout-ms 5000 | jq -r '.type + " | " + .module + " | " + .msg' || echo "No messages found"

echo -e "\n${YELLOW}Step 5: Checking Elasticsearch indices${NC}"
echo "Indices created:"
curl -s -X GET "localhost:9200/_cat/indices/logharbour-*?v&s=index" | grep -v "\.kibana"

echo -e "\n${YELLOW}Step 6: Querying logs from Elasticsearch${NC}"

# Function to query Elasticsearch
query_logs() {
    local log_type=$1
    local description=$2
    
    echo -e "\n${BLUE}$description${NC}"
    curl -s -X GET "localhost:9200/logharbour-${log_type}-*/_search?size=3&sort=when:desc" \
        -H 'Content-Type: application/json' \
        -d '{
            "query": {"match_all": {}},
            "_source": ["type", "module", "msg", "who", "pri"]
        }' | jq -r '.hits.hits[]._source | "\(.type) | \(.module) | \(.msg)"' || echo "No logs found"
}

query_logs "a" "Recent activity logs"
query_logs "c" "Recent change logs"
query_logs "d" "Recent debug logs"

echo -e "\n${YELLOW}Step 7: Consumer status${NC}"
echo "Note: LogHarbour consumer not included in current docker-compose setup"
echo "Logs would be written to Kafka but not consumed to Elasticsearch without the consumer"

echo -e "\n${YELLOW}Step 8: Cleanup${NC}"
echo "Stopping application..."
kill $APP_PID 2>/dev/null
wait $APP_PID 2>/dev/null

echo -e "\n${GREEN}=== Test Complete ===${NC}"
echo -e "\nYou can now:"
echo "1. Access Kafka UI at http://localhost:8090 to view messages"
echo "2. Query Elasticsearch directly at http://localhost:9200"
echo "3. Note: Without LogHarbour consumer, logs won't be in Elasticsearch"
echo -e "\nApplication log saved in: app.log"