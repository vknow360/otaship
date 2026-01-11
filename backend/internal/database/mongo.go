// Package database handles MongoDB connection and operations.
package database

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB holds the database connection and collections.
type MongoDB struct {
	Client    *mongo.Client
	Database  *mongo.Database
	Updates   *mongo.Collection
	Downloads *mongo.Collection
	connected bool
	mu        sync.RWMutex
	APIKeys   *mongo.Collection
}

// Global database instance
var DB *MongoDB

// Config holds MongoDB configuration.
type Config struct {
	URI          string
	DatabaseName string
	Timeout      time.Duration
}

// Connect establishes a connection to MongoDB Atlas.
func Connect(cfg Config) (*MongoDB, error) {
	if cfg.URI == "" {
		log.Println("MongoDB URI not configured, running without database")
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	// Set client options
	clientOptions := options.Client().
		ApplyURI(cfg.URI).
		SetServerAPIOptions(options.ServerAPI(options.ServerAPIVersion1))

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	// Get database and collections
	dbName := cfg.DatabaseName
	if dbName == "" {
		dbName = "otaship"
	}

	db := client.Database(dbName)

	mongodb := &MongoDB{
		Client:    client,
		Database:  db,
		Updates:   db.Collection("updates"),
		Downloads: db.Collection("downloads"),
		APIKeys:   db.Collection("api_keys"),
		connected: true,
	}

	DB = mongodb
	log.Printf("Connected to MongoDB (database: %s)", dbName)

	return mongodb, nil
}

// IsConnected returns true if database is connected.
func (m *MongoDB) IsConnected() bool {
	if m == nil {
		return false
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.connected
}

// Disconnect closes the MongoDB connection.
func (m *MongoDB) Disconnect() error {
	if m == nil || !m.connected {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := m.Client.Disconnect(ctx); err != nil {
		return fmt.Errorf("failed to disconnect from MongoDB: %w", err)
	}

	m.connected = false
	log.Println("Disconnected from MongoDB")
	return nil
}

// HealthCheck verifies the database connection is healthy.
func (m *MongoDB) HealthCheck() error {
	if m == nil || !m.connected {
		return fmt.Errorf("database not connected")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return m.Client.Ping(ctx, nil)
}
