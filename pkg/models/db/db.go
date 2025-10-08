// Package db provides Neo4j database connectivity and query utilities for the
// golf booking application. It manages the database connection lifecycle,
// provides CRUD operations for nodes and relationships, and handles data
// serialization between Go structs and Neo4j graph structures.
//
// The package uses a singleton pattern for the database connection and provides
// thread-safe access to the Neo4j driver.
package db

import (
	"bigfoot/golf/common/helper"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/config"
)

type Database struct {
	Driver neo4j.DriverWithContext
	//mu     sync.RWMutex
	ctx context.Context
	Err error
}

// DynamicNode represents any node that can be saved to Neo4j
type DynamicNode struct {
	Label      string         `json:"label"`
	Properties map[string]any `json:"properties"`
	ID         string         `json:"id,omitempty"`
}

var (
	// Instance is the singleton database connection instance
	Instance     *Database
	once         sync.Once
	// TimeLocation is the configured timezone for the application
	TimeLocation *time.Location
)

// InitDB initializes the Neo4j database connection using environment variables.
// It uses a singleton pattern to ensure only one connection is created.
// The connection is verified before being made available.
//
// Required environment variables:
// - DB_ADMIN: Neo4j admin password (required)
// - DB_URI: Neo4j connection URI (defaults to bolt://localhost:7687)
// - DB_USER: Neo4j username (defaults to neo4j)
//
// The function is safe to call multiple times - initialization only happens once.
func InitDB(ctx context.Context) {

	once.Do(func() {
		dbURI := helper.GetEnvOrDefault("DB_URI", helper.DefaultDBURI)
		dbUser := helper.GetEnvOrDefault("DB_USER", helper.DefaultDBUser)
		dbPassword := os.Getenv("DB_ADMIN")
		Instance = &Database{}
		var err error
		fmt.Println("Driver Connection: ", dbURI, dbPassword, dbUser)
		Instance.Driver, err = neo4j.NewDriverWithContext(
			dbURI,
			neo4j.BasicAuth(dbUser, dbPassword, ""),
			func(c *config.Config) {
				// Configure connection pool settings
				c.MaxConnectionPoolSize = helper.DefaultDBPoolSize
				c.ConnectionAcquisitionTimeout = helper.DefaultHTTPTimeout
			})

		if err != nil {
			Instance.Err = err
		}

		err = Instance.Driver.VerifyConnectivity(ctx)
		if err != nil {
			Instance.Err = err
		}
		Instance.ctx = ctx
		fmt.Println("Connection established.")

		//defer Neo.session.Close(ctx)
	})
}

func (db *Database) NewWriteSession(ctx context.Context) neo4j.SessionWithContext {
	return db.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
}

func (db *Database) NewReadSession(ctx context.Context) neo4j.SessionWithContext {
	return db.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
}

// Save Dynamic Node
func (db *Database) SaveDynamicNode(nd DynamicNode) (string, error) {
	return db.saveNode(db.ctx, nd.Label, nd.Properties)
}

// Query nodes with their relationships
func (db *Database) QueryForJSON(query string, params map[string]any) ([]byte, error) {
	session := db.Driver.NewSession(db.ctx, neo4j.SessionConfig{})
	defer session.Close(db.ctx)

	result, err := session.Run(db.ctx, query, params)
	if err != nil {
		return nil, err
	}

	// Parse results
	var response []map[string]any
	for result.Next(db.ctx) {
		record := result.Record()
		nodeData, _ := record.Get("data") // record.Get("data")
		//response = append(response, nodeData["data"])
		if dataMap, ok := nodeData.(neo4j.Node); ok { //map[string]any); ok {
			response = append(response, dataMap.Props)
		} else if jsonMap, ok := nodeData.(map[string]any); ok {
			response = append(response, jsonMap)
		}
	}
	if len(response) < 1 {
		return nil, nil
	}
	// Convert to JSON first
	jsonData, err := json.Marshal(response)
	return jsonData, err

}

// Query nodes with their relationships
func (db *Database) QueryForMap(query string, params map[string]any) ([]map[string]any, error) {
	session := db.Driver.NewSession(db.ctx, neo4j.SessionConfig{})
	defer session.Close(db.ctx)

	result, err := session.Run(db.ctx, query, params)
	if err != nil {
		return nil, err
	}

	// Parse results
	var response []map[string]any
	for result.Next(db.ctx) {
		record := result.Record()
		nodeData, _ := record.Get("data") // record.Get("data")
		//response = append(response, nodeData["data"])
		if dataMap, ok := nodeData.(neo4j.Node); ok { //map[string]any); ok {
			response = append(response, dataMap.Props)
		} else if jsonMap, ok := nodeData.(map[string]any); ok {
			response = append(response, jsonMap)
		}
	}
	if len(response) < 1 {
		return nil, nil
	}
	return response, err

}
