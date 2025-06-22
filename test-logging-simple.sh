#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Simple LogHarbour Pipeline Test ===${NC}\n"

# First, ensure database migrations are run
echo -e "${YELLOW}Running database migrations...${NC}"
migrate -path migrations -database "postgres://alyatest:alyatest@localhost:5432/alyatest?sslmode=disable" up

echo -e "\n${YELLOW}Starting the application...${NC}"
go run . > app.log 2>&1 &
APP_PID=$!
echo "Application PID: $APP_PID"

# Wait for application to start
sleep 5

echo -e "\n${YELLOW}Testing API endpoints...${NC}"

# Create a user
echo -e "\n${BLUE}1. Creating a new user${NC}"
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "username": "johndoe",
    "phone_number": "+1234567890"
  }' | jq .

sleep 2

# Update the user
echo -e "\n${BLUE}2. Updating the user${NC}"
curl -X POST http://localhost:8080/users/update \
  -H "Content-Type: application/json" \
  -d '{
    "id": 1,
    "name": "John Smith",
    "email": "john.smith@example.com"
  }' | jq .

sleep 2

# Try to update with invalid data
echo -e "\n${BLUE}3. Testing validation error${NC}"
curl -X POST http://localhost:8080/users/update \
  -H "Content-Type: application/json" \
  -d '{
    "id": 1,
    "email": "invalid-email"
  }' | jq .

sleep 2

# Check Kafka messages
echo -e "\n${YELLOW}Checking Kafka messages...${NC}"
echo "Last 10 messages:"
docker exec alyatest-kafka kafka-console-consumer \
    --bootstrap-server localhost:9092 \
    --topic logharbour-logs \
    --max-messages 10 \
    --from-beginning \
    --timeout-ms 5000 2>/dev/null | jq -c '{type, module, msg, pri}' || echo "Failed to read Kafka messages"

# Check Elasticsearch
echo -e "\n${YELLOW}Checking Elasticsearch indices...${NC}"
curl -s "localhost:9200/_cat/indices/logharbour-*?v" | grep -v "\.kibana"

echo -e "\n${YELLOW}Querying activity logs from Elasticsearch...${NC}"
curl -s -X GET "localhost:9200/logharbour-a-*/_search?size=5" \
  -H 'Content-Type: application/json' \
  -d '{
    "query": {"match_all": {}},
    "sort": [{"when": {"order": "desc"}}],
    "_source": ["type", "module", "msg", "who", "pri", "when"]
  }' | jq '.hits.hits[]._source'

echo -e "\n${YELLOW}Querying change logs from Elasticsearch...${NC}"
curl -s -X GET "localhost:9200/logharbour-c-*/_search?size=5" \
  -H 'Content-Type: application/json' \
  -d '{
    "query": {"match_all": {}},
    "sort": [{"when": {"order": "desc"}}],
    "_source": ["type", "module", "msg", "data.change_data"]
  }' | jq '.hits.hits[]._source'

# Consumer status
echo -e "\n${YELLOW}Consumer status:${NC}"
docker logs --tail 5 alyatest-logharbour-consumer

# Cleanup
echo -e "\n${YELLOW}Stopping application...${NC}"
kill $APP_PID 2>/dev/null
wait $APP_PID 2>/dev/null

echo -e "\n${GREEN}Test complete!${NC}"
echo "Check app.log for application logs"