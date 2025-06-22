package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

type LogEntry struct {
	ID       string                 `json:"id"`
	App      string                 `json:"app"`
	System   string                 `json:"system"`
	Module   string                 `json:"module,omitempty"`
	Type     string                 `json:"type"`
	Priority string                 `json:"pri"`
	When     string                 `json:"when"`
	Who      string                 `json:"who,omitempty"`
	RemoteIP string                 `json:"remote_ip,omitempty"`
	TraceID  string                 `json:"trace_id,omitempty"`
	Msg      string                 `json:"msg"`
	Data     map[string]interface{} `json:"data,omitempty"`
}

func main() {
	// Get configuration from environment
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:9092"
	}
	
	elasticsearchURL := os.Getenv("ELASTICSEARCH_URL")
	if elasticsearchURL == "" {
		elasticsearchURL = "http://localhost:9200"
	}

	// Kafka configuration
	brokers := strings.Split(kafkaBrokers, ",")
	topic := "logharbour-logs"
	group := "logharbour-consumer"

	// Elasticsearch configuration
	cfg := elasticsearch.Config{
		Addresses: []string{elasticsearchURL},
	}

	// Create Elasticsearch client
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatalf("Error creating Elasticsearch client: %s", err)
	}

	// Test Elasticsearch connection
	res, err := es.Info()
	if err != nil {
		log.Fatalf("Error getting Elasticsearch info: %s", err)
	}
	defer res.Body.Close()
	log.Println("Elasticsearch connected successfully")

	// Create index template for logs
	createIndexTemplate(es)

	// Kafka consumer configuration
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Version = sarama.V2_6_0_0

	// Create consumer group
	consumerGroup, err := sarama.NewConsumerGroup(brokers, group, config)
	if err != nil {
		log.Fatalf("Error creating consumer group: %s", err)
	}
	defer consumerGroup.Close()

	// Create consumer handler
	consumer := &Consumer{
		es: es,
	}

	// Setup signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, os.Interrupt)

	// Start consuming
	go func() {
		for {
			if err := consumerGroup.Consume(ctx, []string{topic}, consumer); err != nil {
				log.Printf("Error from consumer: %v", err)
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()

	log.Println("LogHarbour consumer started. Press Ctrl+C to exit.")
	<-sigterm
	log.Println("Shutting down consumer...")
}

// Consumer represents a Sarama consumer group consumer
type Consumer struct {
	es *elasticsearch.Client
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages()
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		// Parse log entry
		var logEntry LogEntry
		if err := json.Unmarshal(message.Value, &logEntry); err != nil {
			log.Printf("Error parsing message: %v", err)
			session.MarkMessage(message, "")
			continue
		}

		// Index to Elasticsearch
		if err := consumer.indexLog(logEntry); err != nil {
			log.Printf("Error indexing log: %v", err)
			// Continue processing other messages even if one fails
		}

		session.MarkMessage(message, "")
	}
	return nil
}

func (consumer *Consumer) indexLog(logEntry LogEntry) error {
	// Determine index name based on log type and date
	indexName := fmt.Sprintf("logharbour-%s-%s", 
		strings.ToLower(logEntry.Type), 
		time.Now().Format("2006.01.02"))

	// Convert log entry to JSON
	data, err := json.Marshal(logEntry)
	if err != nil {
		return fmt.Errorf("error marshaling log entry: %w", err)
	}

	// Create index request
	req := esapi.IndexRequest{
		Index:      indexName,
		DocumentID: logEntry.ID,
		Body:       strings.NewReader(string(data)),
		Refresh:    "false",
	}

	// Perform the request
	res, err := req.Do(context.Background(), consumer.es)
	if err != nil {
		return fmt.Errorf("error indexing document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error indexing document: %s", res.String())
	}

	return nil
}

func createIndexTemplate(es *elasticsearch.Client) {
	// Create an index template for LogHarbour logs
	template := `{
		"index_patterns": ["logharbour-*"],
		"template": {
			"settings": {
				"number_of_shards": 1,
				"number_of_replicas": 0
			},
			"mappings": {
				"properties": {
					"id": { "type": "keyword" },
					"app": { "type": "keyword" },
					"system": { "type": "keyword" },
					"module": { "type": "keyword" },
					"type": { "type": "keyword" },
					"pri": { "type": "keyword" },
					"when": { "type": "date" },
					"who": { "type": "keyword" },
					"remote_ip": { "type": "ip" },
					"trace_id": { "type": "keyword" },
					"msg": { "type": "text" },
					"data": { "type": "object" }
				}
			}
		}
	}`

	req := esapi.IndicesPutIndexTemplateRequest{
		Name: "logharbour-template",
		Body: strings.NewReader(template),
	}

	res, err := req.Do(context.Background(), es)
	if err != nil {
		log.Printf("Error creating index template: %s", err)
		return
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Printf("Error creating index template: %s", res.String())
	} else {
		log.Println("Index template created successfully")
	}
}