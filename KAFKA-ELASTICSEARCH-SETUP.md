# Kafka and Elasticsearch Setup for LogHarbour

This document explains how LogHarbour is configured to write logs to Kafka and how a consumer service indexes them into Elasticsearch.

## Architecture Overview

```
Application → LogHarbour → Kafka → Consumer → Elasticsearch → Kibana
                  ↓
               Stdout (fallback)
```

## Components

### 1. Kafka
- **Purpose**: Message broker for reliable log delivery
- **Topic**: `logharbour-logs`
- **Port**: 9092 (external), 29092 (internal)
- **Configuration**: Auto-create topics enabled

### 2. Elasticsearch
- **Purpose**: Log storage and search engine
- **Port**: 9200 (REST API), 9300 (transport)
- **Index Pattern**: `logharbour-{type}-{date}`
  - Example: `logharbour-a-2024.06.22` for activity logs
  - Types: `a` (activity), `c` (change), `d` (debug)

### 3. Kibana
- **Purpose**: Log visualization and analysis
- **Port**: 5601
- **URL**: http://localhost:5601

### 4. LogHarbour Consumer
- **Purpose**: Consumes logs from Kafka and indexes to Elasticsearch
- **Consumer Group**: `logharbour-consumer`
- **Features**:
  - Automatic index creation
  - Index template management
  - Error resilience

## Setup Instructions

### 1. Start Infrastructure
```bash
docker-compose up -d
```

This starts:
- PostgreSQL (existing)
- Redis (existing)
- etcd (existing)
- Zookeeper (new)
- Kafka (new)
- Elasticsearch (new)
- Kibana (new)
- LogHarbour Consumer (new)

### 2. Verify Services
```bash
# Check all services are running
docker-compose ps

# Check Kafka topics
docker exec -it alyatest-kafka kafka-topics --bootstrap-server localhost:9092 --list

# Check Elasticsearch health
curl -X GET "localhost:9200/_cat/health?v"

# Check indices
curl -X GET "localhost:9200/_cat/indices?v"
```

### 3. Application Configuration

The application (main.go) is configured to:
1. Try to connect to Kafka
2. If successful, use Kafka as primary writer with stdout as fallback
3. If Kafka is unavailable, use stdout only

```go
// Kafka configuration in main.go
kafkaConfig := logharbour.KafkaConfig{
    Brokers: []string{"localhost:9092"},
    Topic:   "logharbour-logs",
}
```

## Log Types and Indices

### Activity Logs (Type: A)
- Index: `logharbour-a-YYYY.MM.DD`
- Contains: API calls, user actions, system events

### Change Logs (Type: C)
- Index: `logharbour-c-YYYY.MM.DD`
- Contains: Data modifications, audit trail

### Debug Logs (Type: D)
- Index: `logharbour-d-YYYY.MM.DD`
- Contains: HTTP requests, debug information

## Viewing Logs in Kibana

1. Access Kibana: http://localhost:5601
2. Go to "Stack Management" → "Index Patterns"
3. Create index pattern: `logharbour-*`
4. Set timestamp field: `when`
5. Go to "Discover" to view logs

### Sample Queries

```
# View all activity logs
type: "A"

# View logs for specific user
who: "admin@example.com"

# View logs for specific module
module: "UserService"

# View error logs
pri: "Err" OR pri: "Crit"

# View data changes for users
type: "C" AND data.change_data.entity: "User"
```

## Monitoring

### Kafka Monitoring
```bash
# View consumer group status
docker exec -it alyatest-kafka kafka-consumer-groups \
  --bootstrap-server localhost:9092 \
  --group logharbour-consumer \
  --describe

# View topic details
docker exec -it alyatest-kafka kafka-topics \
  --bootstrap-server localhost:9092 \
  --topic logharbour-logs \
  --describe
```

### Elasticsearch Monitoring
```bash
# Cluster health
curl -X GET "localhost:9200/_cluster/health?pretty"

# Index stats
curl -X GET "localhost:9200/logharbour-*/_stats?pretty"

# Check consumer logs
docker logs alyatest-logharbour-consumer
```

## Troubleshooting

### Logs not appearing in Elasticsearch
1. Check if Kafka is receiving messages:
   ```bash
   docker exec -it alyatest-kafka kafka-console-consumer \
     --bootstrap-server localhost:9092 \
     --topic logharbour-logs \
     --from-beginning
   ```

2. Check consumer logs:
   ```bash
   docker logs alyatest-logharbour-consumer
   ```

3. Check Elasticsearch logs:
   ```bash
   docker logs alyatest-elasticsearch
   ```

### Consumer not starting
- Ensure Kafka and Elasticsearch are healthy
- Check consumer can connect to both services
- Verify network connectivity between containers

### High memory usage
- Elasticsearch is configured with 512MB heap
- Adjust in docker-compose.yaml if needed:
  ```yaml
  environment:
    - "ES_JAVA_OPTS=-Xms1g -Xmx1g"
  ```

## Production Considerations

1. **Kafka Configuration**:
   - Increase replication factor
   - Configure retention policies
   - Set up multiple brokers

2. **Elasticsearch Configuration**:
   - Enable security (xpack.security.enabled=true)
   - Configure snapshots for backup
   - Set up proper index lifecycle management

3. **Consumer Configuration**:
   - Run multiple consumer instances
   - Configure proper error handling
   - Add metrics and monitoring

4. **LogHarbour Configuration**:
   - Tune connection pool size
   - Configure timeouts appropriately
   - Monitor fallback writer usage