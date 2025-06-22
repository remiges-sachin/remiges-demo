package pg

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/synapsewave/remiges-demo/pg/sqlc-gen"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

type Provider struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

func NewProvider(cfg Config) *Provider {
	ctx := context.Background()
	
	// Build connection string in pgx format
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", 
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	// Create a connection pool
	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		log.Fatalf("Failed to parse connection string: %v", err)
	}
	
	// You can configure pool settings here if needed
	// poolConfig.MaxConns = 25
	// poolConfig.MinConns = 5
	
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatalf("Failed to create connection pool: %v", err)
	}

	// Test the connection
	err = pool.Ping(ctx)
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Successfully connected to the database using pgxpool")

	queries := sqlc.New(pool)

	return &Provider{pool: pool, queries: queries}
}

func (p *Provider) Pool() *pgxpool.Pool {
	return p.pool
}

func (p *Provider) Queries() *sqlc.Queries {
	return p.queries
}

func (p *Provider) Close() {
	p.pool.Close()
}