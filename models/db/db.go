package db

import (
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
	driver neo4j.DriverWithContext
	//mu     sync.RWMutex
	ctx context.Context
}

// DynamicNode represents any node that can be saved to Neo4j
type DynamicNode struct {
	Label      string         `json:"label"`
	Properties map[string]any `json:"properties"`
	ID         string         `json:"id,omitempty"`
}

var (
	Instance *Database
	once     sync.Once
)

func InitDB(ctx context.Context) {

	once.Do(func() {
		//
		dbURI := "bolt://localhost:7687"
		dbUser := "neo4j"
		dbPassword := os.Getenv("DB_ADMIN")
		Instance = &Database{}
		var err error
		Instance.driver, err = neo4j.NewDriverWithContext(
			dbURI,
			neo4j.BasicAuth(dbUser, dbPassword, ""),
			func(c *config.Config) {
				// Optional: Configure connection pool settings
				c.MaxConnectionPoolSize = 50
				c.ConnectionAcquisitionTimeout = time.Second * 30 // seconds
			})

		if err != nil {
			panic(err)
		}

		err = Instance.driver.VerifyConnectivity(ctx)
		if err != nil {
			panic(err)
		}
		Instance.ctx = ctx
		fmt.Println("Connection established.")

		//defer Neo.session.Close(ctx)
	})

}

func (db *Database) NewWriteSession(ctx context.Context) neo4j.SessionWithContext {
	return db.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
}

func (db *Database) NewReadSession(ctx context.Context) neo4j.SessionWithContext {
	return db.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
}

// Save Dynamic Node
func (db *Database) SaveDynamicNode(nd DynamicNode) (string, error) {
	return db.saveNode(db.ctx, nd.Label, nd.Properties)
}

// Query nodes with their relationships
func (db *Database) QueryForJSON(query string, params map[string]any) ([]byte, error) {
	session := db.driver.NewSession(db.ctx, neo4j.SessionConfig{})
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
	session := db.driver.NewSession(db.ctx, neo4j.SessionConfig{})
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
