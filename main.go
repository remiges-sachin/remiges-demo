package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/synapsewave/remiges-demo/pg"
	usersvc "github.com/synapsewave/remiges-demo/userservice"
	"github.com/remiges-tech/alya/router"
	"github.com/remiges-tech/alya/service"
	"github.com/remiges-tech/logharbour/logharbour"

	"github.com/remiges-tech/rigel"
	"github.com/remiges-tech/rigel/etcd"
)

func main() {
	// Initialize logger context
	lctx := logharbour.NewLoggerContext(logharbour.DefaultPriority)
	
	// Initialize Kafka writer for LogHarbour
	kafkaConfig := logharbour.KafkaConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "logharbour-logs",
	}
	
	// Create Kafka writer with connection pool
	kafkaWriter, err := logharbour.NewKafkaWriter(kafkaConfig, logharbour.WithPoolSize(10))
	
	// Initialize logger with appropriate writer
	var logger *logharbour.Logger
	if err != nil {
		// If Kafka is not available, use stdout only
		fmt.Printf("Warning: Failed to create Kafka writer: %v. Using stdout only.\n", err)
		fallbackWriter := logharbour.NewFallbackWriter(os.Stdout, os.Stdout)
		logger = logharbour.NewLogger(lctx, "UserService", fallbackWriter)
	} else {
		// Create fallback writer (Kafka primary, stdout fallback)
		defer kafkaWriter.Close()
		fallbackWriter := logharbour.NewFallbackWriter(kafkaWriter, os.Stdout)
		logger = logharbour.NewLogger(lctx, "UserService", fallbackWriter)
	}
	logger.WithPriority(logharbour.Debug2)

	// Initialize etcd storage for Rigel
	etcdEndpoints := []string{"localhost:2379"}
	etcdStorage, err := etcd.NewEtcdStorage(etcdEndpoints)
	if err != nil {
		logger.Error(fmt.Errorf("failed to create EtcdStorage: %w", err)).LogActivity("Startup failed", nil)
		os.Exit(1)
	}

	// Initialize Rigel client
	rigelClient := rigel.New(etcdStorage, "alya", "usersvc", 1, "dev")
	logger.Info().LogActivity("Rigel client initialized", nil)

	// Create context
	ctx := context.Background()

	// Get database configuration dynamically
	host, err := rigelClient.Get(ctx, "database.host")
	if err != nil {
		logger.Error(fmt.Errorf("failed to get database host: %w", err)).LogActivity("Configuration error", nil)
		os.Exit(1)
	}
	port, err := rigelClient.GetInt(ctx, "database.port")
	if err != nil {
		logger.Error(fmt.Errorf("failed to get database port: %w", err)).LogActivity("Configuration error", nil)
		os.Exit(1)
	}
	user, err := rigelClient.Get(ctx, "database.user")
	if err != nil {
		logger.Error(fmt.Errorf("failed to get database user: %w", err)).LogActivity("Configuration error", nil)
		os.Exit(1)
	}
	password, err := rigelClient.Get(ctx, "database.password")
	if err != nil {
		logger.Error(fmt.Errorf("failed to get database password: %w", err)).LogActivity("Configuration error", nil)
		os.Exit(1)
	}
	dbname, err := rigelClient.Get(ctx, "database.dbname")
	if err != nil {
		logger.Error(fmt.Errorf("failed to get database name: %w", err)).LogActivity("Configuration error", nil)
		os.Exit(1)
	}

	// Initialize database
	dbConfig := pg.Config{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbname,
	}
	provider := pg.NewProvider(dbConfig)
	db := provider.Queries()
	logger.Info().LogActivity("Database connection initialized", map[string]any{
		"host": host,
		"port": port,
		"user": user,
		"db":   dbname,
	})

	// Create LogHarbour adapter for request logging
	// This enables automatic logging of all HTTP requests with comprehensive details
	logAdapter := router.NewLogHarbourAdapter(logger)

	// Create Gin router with middleware
	// Note: Using gin.New() instead of gin.Default() to have full control over middleware
	r := gin.New()
	r.Use(gin.Recovery()) // Recover from panics and return 500 error
	r.Use(router.LogRequest(logAdapter)) // Log all HTTP requests automatically

	// Create service with Rigel client
	s := service.NewService(r).
		WithLogHarbour(logger).
		WithDatabase(db).
		WithRigelConfig(rigelClient)

	// Register routes
	s.RegisterRoute("POST", "/users", usersvc.HandleCreateUserRequest)
	s.RegisterRoute("POST", "/users/update", usersvc.HandleUpdateUserRequest)
	logger.Info().LogActivity("Routes registered", nil)

	// Get server port dynamically
	serverPort, err := rigelClient.GetInt(ctx, "server.port")
	if err != nil {
		logger.Error(fmt.Errorf("failed to get server port: %w", err)).LogActivity("Configuration error", nil)
		os.Exit(1)
	}

	// Start server
	serverAddr := fmt.Sprintf(":%d", serverPort)
	logger.Info().LogActivity("Starting server", map[string]any{"port": serverPort})
	if err := r.Run(serverAddr); err != nil {
		logger.Error(fmt.Errorf("error starting server: %w", err)).LogActivity("Server startup failed", nil)
		os.Exit(1)
	}
}
