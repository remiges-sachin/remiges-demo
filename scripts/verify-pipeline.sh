#!/bin/bash

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}=== LogHarbour Pipeline Verification ===${NC}\n"

# Check services
echo -e "${YELLOW}Service Status:${NC}"
echo -n "PostgreSQL: "
docker exec demo-postgres pg_isready -U remiges >/dev/null 2>&1 && echo -e "${GREEN}✓${NC}" || echo "✗"

echo -n "etcd: "
curl -s http://localhost:2379/version >/dev/null 2>&1 && echo -e "${GREEN}✓${NC}" || echo "✗"

echo -n "Kafka: "
docker exec demo-kafka kafka-broker-api-versions --bootstrap-server localhost:9092 >/dev/null 2>&1 && echo -e "${GREEN}✓${NC}" || echo "✗"

echo -n "Elasticsearch: "
curl -s http://localhost:9200/_cat/health >/dev/null 2>&1 && echo -e "${GREEN}✓${NC}" || echo "✗"

echo -n "Kibana: "
curl -s http://localhost:5601/api/status >/dev/null 2>&1 && echo -e "${GREEN}✓${NC}" || echo "✗"

echo -n "Kafka UI: "
curl -s http://localhost:8090 >/dev/null 2>&1 && echo -e "${GREEN}✓${NC}" || echo "✗"

echo -n "LogHarbour Consumer: "
docker ps --format "table {{.Names}}\t{{.Status}}" | grep -q "demo-logharbour-consumer.*Up" && echo -e "${GREEN}✓${NC}" || echo "✗"

# Show stats
echo -e "\n${YELLOW}Pipeline Statistics:${NC}"

# Kafka messages
KAFKA_COUNT=$(docker exec demo-kafka kafka-run-class kafka.tools.GetOffsetShell \
    --broker-list localhost:9092 \
    --topic logharbour-logs 2>/dev/null | awk -F: '{sum += $3} END {print sum}')
echo "Kafka messages in topic: ${KAFKA_COUNT:-0}"

# Elasticsearch documents
ES_COUNT=$(curl -s "localhost:9200/logharbour-*/_count" | jq '.count' 2>/dev/null)
echo "Elasticsearch documents: ${ES_COUNT:-0}"

# Breakdown by type
echo -e "\n${YELLOW}Logs by Type:${NC}"
for type in a c d; do
    count=$(curl -s "localhost:9200/logharbour-$type-*/_count" 2>/dev/null | jq '.count' 2>/dev/null || echo "0")
    case $type in
        a) echo "Activity logs (A): $count" ;;
        c) echo "Change logs (C): $count" ;;
        d) echo "Debug logs (D): $count" ;;
    esac
done

# Recent logs
echo -e "\n${YELLOW}Most Recent Logs:${NC}"
curl -s -X GET "localhost:9200/logharbour-*/_search" \
  -H 'Content-Type: application/json' \
  -d '{
    "size": 5,
    "sort": [{"when": {"order": "desc"}}],
    "_source": ["type", "module", "msg", "when"]
  }' 2>/dev/null | jq -r '.hits.hits[]._source | "\(.when | .[0:19]) | \(.type) | \(.module) | \(.msg)"' | column -t -s '|'

echo -e "\n${GREEN}Verification complete!${NC}"
echo -e "\nAccess points:"
echo "• Elasticsearch: http://localhost:9200"
echo "• Kibana: http://localhost:5601"
echo "• Kafka UI: http://localhost:8090"
echo "• LogHarbour Consumer: Running in docker-compose"