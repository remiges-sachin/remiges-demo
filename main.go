package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/synapsewave/remiges-demo/pg"
	usersvc "github.com/synapsewave/remiges-demo/userservice"
	"github.com/remiges-tech/alya/config"
	"github.com/remiges-tech/alya/router"
	"github.com/remiges-tech/alya/service"
	"github.com/remiges-tech/logharbour/logharbour"

	"github.com/remiges-tech/rigel"
	"github.com/remiges-tech/rigel/etcd"
)

// AppConfig holds application configuration from config.json
type AppConfig struct {
	Server ServerConfig `json:"server"`
}

// ServerConfig holds server settings
type ServerConfig struct {
	Port int `json:"port"`
}

func main() {
	// ===== LogHarbour Setup =====
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
		// Create multi-writer to write to both Kafka and stdout using Go's io.MultiWriter
		defer kafkaWriter.Close()
		multiWriter := io.MultiWriter(kafkaWriter, os.Stdout)
		fallbackWriter := logharbour.NewFallbackWriter(multiWriter, os.Stdout)
		logger = logharbour.NewLogger(lctx, "UserService", fallbackWriter)
	}
	logger.WithPriority(logharbour.Debug2)
	
	logger.Info().LogActivity("Starting User Service", nil)

	// ===== Rigel Configuration Setup =====
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

	// ===== Load Server Configuration from File =====
	// Load server configuration from config.json
	var appConfig AppConfig
	err = config.LoadConfigFromFile("config.json", &appConfig)
	if err != nil {
		logger.Error(fmt.Errorf("failed to load config.json: %w", err)).LogActivity("Configuration error", nil)
		os.Exit(1)
	}
	logger.Info().LogActivity("Server configuration loaded from config.json", map[string]any{
		"server_port": appConfig.Server.Port,
	})

	// ===== Database Configuration from Rigel =====
	// Get database configuration from Rigel (etcd)
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

	// ===== Database Initialization =====
	// Initialize database using Rigel configuration
	dbConfig := pg.Config{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbname,
	}
	provider := pg.NewProvider(dbConfig)
	defer provider.Close() // Ensure connection pool is closed on exit
	db := provider.Queries()
	logger.Info().LogActivity("Database connection initialized", map[string]any{
		"host": host,
		"port": port,
		"user": user,
		"db":   dbname,
	})

	// ===== HTTP Router and Middleware Setup =====
	// Create LogHarbour adapter for request logging
	// This enables automatic logging of all HTTP requests with comprehensive details
	logAdapter := router.NewLogHarbourAdapter(logger)

	// Create Gin router with middleware
	// Note: Using gin.New() instead of gin.Default() to have full control over middleware
	r := gin.New()
	r.Use(gin.Recovery()) // Recover from panics and return 500 error
	r.Use(router.LogRequest(logAdapter)) // Log all HTTP requests automatically

	// ===== Alya Service Setup =====
	// Create service with Rigel client
	s := service.NewService(r).
		WithLogHarbour(logger).
		WithDatabase(db).
		WithRigelConfig(rigelClient)

	// Register routes
	s.RegisterRoute("POST", "/user_create", usersvc.HandleCreateUserRequest)
	s.RegisterRoute("POST", "/user_get", usersvc.HandleGetUserRequest)
	s.RegisterRoute("POST", "/user_update", usersvc.HandleUpdateUserRequest)
	logger.Info().LogActivity("Routes registered", nil)

	// ===== Server Configuration and Startup =====
	// Start server using configuration from struct
	serverAddr := fmt.Sprintf(":%d", appConfig.Server.Port)
	logger.Info().LogActivity("Starting server", map[string]any{"port": appConfig.Server.Port})
	if err := r.Run(serverAddr); err != nil {
		logger.Error(fmt.Errorf("error starting server: %w", err)).LogActivity("Server startup failed", nil)
		os.Exit(1)
	}
}
