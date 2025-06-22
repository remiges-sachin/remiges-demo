#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Complete LogHarbour Kafka-Elasticsearch Pipeline Test ===${NC}\n"

# Kill any existing processes on port 8080
echo -e "${YELLOW}Cleaning up any existing processes...${NC}"
lsof -ti:8080 | xargs -r kill -9 2>/dev/null

# Note about consumer
echo -e "${YELLOW}Note: LogHarbour consumer not included in current setup${NC}"
echo -e "${YELLOW}Logs will be written to Kafka but not consumed to Elasticsearch${NC}"

echo -e "\n${YELLOW}Starting the application...${NC}"
go run . > app.log 2>&1 &
APP_PID=$!
echo "Application PID: $APP_PID"

# Wait for application to start
sleep 5

# Check if app started successfully
if ! ps -p $APP_PID > /dev/null; then
    echo -e "${RED}Application failed to start. Check app.log${NC}"
    tail -20 app.log
    exit 1
fi

echo -e "\n${GREEN}✓ Application started successfully${NC}"

# Test endpoints
echo -e "\n${YELLOW}=== Testing API Endpoints ===${NC}"

# 1. Create user - should generate activity log and change log
echo -e "\n${BLUE}1. Creating a new user (generates Activity + Change logs)${NC}"
RESPONSE=$(curl -s -X POST http://localhost:8080/user_create \
  -H "Content-Type: application/json" \
  -d '{
    "data": {
      "name": "John Doe",
      "email": "john.doe@validmail.com",
      "username": "johndoe",
      "phone_number": "+1234567890"
    }
  }')
echo "Response: $RESPONSE" | jq . 2>/dev/null || echo "$RESPONSE"
sleep 2

# 2. Update user - should generate activity log and change log
echo -e "\n${BLUE}2. Updating user (generates Activity + Change logs)${NC}"
RESPONSE=$(curl -s -X POST http://localhost:8080/user_update \
  -H "Content-Type: application/json" \
  -d '{
    "data": {
      "id": 1,
      "name": "John Smith",
      "email": "john.smith@validmail.com"
    }
  }')
echo "Response: $RESPONSE" | jq . 2>/dev/null || echo "$RESPONSE"
sleep 2

# 3. Validation error - should generate activity log with error
echo -e "\n${BLUE}3. Testing validation error (generates Activity log with error)${NC}"
RESPONSE=$(curl -s -X POST http://localhost:8080/user_update \
  -H "Content-Type: application/json" \
  -d '{
    "data": {
      "id": 1,
      "email": "invalid-email-format"
    }
  }')
echo "Response: $RESPONSE" | jq . 2>/dev/null || echo "$RESPONSE"
sleep 2

# 4. Non-existent user - should generate activity log with error
echo -e "\n${BLUE}4. Updating non-existent user (generates Activity log with error)${NC}"
RESPONSE=$(curl -s -X POST http://localhost:8080/user_update \
  -H "Content-Type: application/json" \
  -d '{
    "data": {
      "id": 999,
      "name": "Ghost User"
    }
  }')
echo "Response: $RESPONSE" | jq . 2>/dev/null || echo "$RESPONSE"
sleep 3

# Check the logging pipeline
echo -e "\n${YELLOW}=== Checking Logging Pipeline ===${NC}"

# Check Kafka
echo -e "\n${BLUE}Kafka Topic Messages (last 5):${NC}"
docker exec demo-kafka kafka-console-consumer \
    --bootstrap-server localhost:9092 \
    --topic logharbour-logs \
    --from-beginning \
    --max-messages 5 \
    --timeout-ms 3000 2>/dev/null | while read line; do
    echo "$line" | jq -c '. | {type, module, msg, pri}' 2>/dev/null || echo "$line"
done

# Check Elasticsearch indices
echo -e "\n${BLUE}Elasticsearch Indices:${NC}"
curl -s "localhost:9200/_cat/indices/logharbour-*?v&h=index,docs.count,store.size" | column -t

# Query logs by type
echo -e "\n${BLUE}Activity Logs (Type A) - Last 3:${NC}"
curl -s -X GET "localhost:9200/logharbour-a-*/_search" \
  -H 'Content-Type: application/json' \
  -d '{
    "size": 3,
    "sort": [{"when": {"order": "desc"}}],
    "query": {"match_all": {}},
    "_source": ["module", "msg", "pri", "who", "when"]
  }' | jq -r '.hits.hits[]._source | "\(.when) | \(.module) | \(.pri) | \(.msg)"' 2>/dev/null

echo -e "\n${BLUE}Change Logs (Type C) - Last 3:${NC}"
curl -s -X GET "localhost:9200/logharbour-c-*/_search" \
  -H 'Content-Type: application/json' \
  -d '{
    "size": 3,
    "sort": [{"when": {"order": "desc"}}],
    "query": {"match_all": {}},
    "_source": ["module", "msg", "data.change_data", "when"]
  }' | jq -r '.hits.hits[] | "\(.when) | \(._source.module) | \(._source.msg) | Entity: \(._source.data.change_data.entity) | ID: \(._source.data.change_data.key)"' 2>/dev/null

# Check for errors
echo -e "\n${BLUE}Error Logs (Priority: Err or Crit):${NC}"
curl -s -X GET "localhost:9200/logharbour-*/_search" \
  -H 'Content-Type: application/json' \
  -d '{
    "size": 5,
    "query": {
      "terms": {
        "pri": ["Err", "Crit"]
      }
    },
    "_source": ["module", "msg", "pri", "when"]
  }' | jq -r '.hits.hits[]._source | "\(.when) | \(.pri) | \(.module) | \(.msg)"' 2>/dev/null

# Consumer status
echo -e "\n${BLUE}Consumer Service Status:${NC}"
echo "Note: LogHarbour consumer not included in current docker-compose setup"

# Summary
echo -e "\n${YELLOW}=== Summary ===${NC}"
echo -e "Total logs in Elasticsearch:"
curl -s -X GET "localhost:9200/logharbour-*/_count" | jq '.count' 2>/dev/null

echo -e "\n${GREEN}✓ Pipeline test complete!${NC}"
echo -e "\nYou can access:"
echo "- Kafka UI: http://localhost:8090"
echo "- Elasticsearch: http://localhost:9200"
echo "- Application logs: cat app.log"
echo "\nNote: Without LogHarbour consumer, logs are in Kafka but not Elasticsearch"

# Cleanup
echo -e "\n${YELLOW}Stopping application...${NC}"
kill $APP_PID 2>/dev/null
wait $APP_PID 2>/dev/null

echo -e "\n${GREEN}Done!${NC}"