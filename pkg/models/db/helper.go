package db

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// Existing helper functions (keeping them from previous version)

func (m *Database) SaveStruct(data interface{}, label string) (string, error) {
	properties := structToMap(label, data)

	return m.saveNode(m.ctx, label, properties)
}

type Relation struct {
	NodeN    string
	NodeX    string
	NodeNID  string
	NodeXID  string
	Name     string
	Property string
	Body     string
}

func (m *Database) SaveRelationship(data Relation) error {

	var _prop string
	if data.Property != "" && data.Body != "" {
		_prop = fmt.Sprintf(`{%s: "%s"}`, data.Property, data.Body)
	}

	_query := fmt.Sprintf(`
		MATCH (n:%s {id: "%s"})
		MATCH (x:%s {id:"%s"})
		MERGE (n)-[r:%s %s]->(x)
		RETURN r`, data.NodeN, data.NodeNID, data.NodeX, data.NodeXID, data.Name, _prop)
	session := m.NewWriteSession(m.ctx)
	_, err := session.ExecuteWrite(m.ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		res, err := tx.Run(m.ctx, _query, nil)
		if err != nil {
			return nil, err
		}
		if res.Next(m.ctx) {
			return res.Record().Values[0], nil
		}
		return nil, res.Err()
	})

	if err != nil {
		return err
	}

	return nil
}

func (m *Database) SaveFromJSON(ctx context.Context, jsonData string, label string) (string, error) {
	var properties map[string]interface{}
	if err := json.Unmarshal([]byte(jsonData), &properties); err != nil {
		return "", err
	}
	return m.saveNode(ctx, label, properties)
}

func (m *Database) saveNode(ctx context.Context, label string, properties map[string]interface{}) (string, error) {

	cleanProps := prepareProperties(properties)
	var cypher string
	var params map[string]interface{}
	if _id, exists := cleanProps["id"]; !exists || _id == "" {
		cleanProps["id"] = generateUUID()
		cypher, params = buildCreateQuery(label, cleanProps)
	} else {
		cypher, params = buildUpdateQuery(label, cleanProps)
	}

	session := m.NewWriteSession(ctx)
	result, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		res, err := tx.Run(ctx, cypher, params)
		if err != nil {
			return nil, err
		}
		if res.Next(ctx) {
			return res.Record().Values[0], nil
		}
		return nil, res.Err()
	})

	if err != nil {
		return "", err
	}

	return result.(string), nil
}

func (m *Database) QueryNodes(label string, filters map[string]interface{}) ([]map[string]any, error) {
	session := m.driver.NewSession(m.ctx, neo4j.SessionConfig{})
	defer session.Close(m.ctx)

	cypher, params := buildQueryCypher(label, filters)
	var nodes []map[string]any

	result, err := session.ExecuteRead(m.ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		res, err := tx.Run(m.ctx, cypher, params)
		if err != nil {
			return nil, err
		}

		for res.Next(m.ctx) {
			record := res.Record()
			if len(record.Values) > 0 {
				if node, ok := record.Values[0].(neo4j.Node); ok {
					nodes = append(nodes, node.Props)
				}
			}
		}
		return nodes, res.Err()
	})

	if err != nil {
		return nil, err
	}

	return result.([]map[string]any), nil
}

// Helper functions from previous version
func structToMap(lbl string, obj interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	fmt.Println("structure for ", lbl)
	//relationships := make(map[string][]string) //label and field names
	v := reflect.ValueOf(obj)
	t := reflect.TypeOf(obj)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		if !field.CanInterface() {
			continue
		}

		jsonTag := fieldType.Tag.Get("json")
		if jsonTag == "" {
			continue
		}
		fieldName := fieldType.Name
		if jsonTag != "" && jsonTag != "-" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" {
				fieldName = parts[0]
			}
		}

		value := field.Interface()
		_t := reflect.TypeOf(value)
		switch _t.Kind() {
		case reflect.Ptr:
			elem := _t.Elem()
			if strings.Contains(elem.PkgPath(), "birdsfoot") {
				continue
			}
		case reflect.Slice, reflect.Array:
			elem := _t.Elem()
			if strings.Contains(elem.PkgPath(), "birdsfoot") {
				continue
			}
		case reflect.Map:
			elemType := _t.Elem()
			if strings.Contains(elemType.PkgPath(), "birdsfoot") {
				continue
			}
		}

		result[fieldName] = value
	}

	return result
}

func prepareProperties(props map[string]interface{}) map[string]interface{} {
	cleaned := make(map[string]interface{})

	for key, value := range props {

		switch v := value.(type) {
		case time.Time:
			if _timeV, ok := value.(time.Time); ok {
				if _timeV.Location() == nil || _timeV.Location().String() == "" {
					_timeV = assignLocation(_timeV)
				}
				cleaned[key] = _timeV
			}
		case []interface{}:
			cleaned[key] = convertSlice(v)
		case map[string]interface{}:
			jsonBytes, _ := json.Marshal(v)
			cleaned[key] = string(jsonBytes)
			fmt.Println("MAP FOUND IN SAVE: ", jsonBytes)
			//continue
		case nil:
			continue
		default:
			t := reflect.TypeOf(value)
			if t.Kind() == reflect.Map {
				continue
			}
			cleaned[key] = value
		}
	}

	return cleaned
}
func assignLocation(myTime time.Time) time.Time {
	return time.Date(myTime.Year(), myTime.Month(), myTime.Day(), myTime.Hour(), myTime.Minute(), myTime.Second(), 0, TimeLocation)
}

/*
	func isArrayOfCustomStructs(obj interface{}) bool {
		//v := reflect.ValueOf(obj)
		t := reflect.TypeOf(obj)

		// Check if it's a slice or array
		if t.Kind() != reflect.Slice && t.Kind() != reflect.Array {
			return false
		}

		// Check if the element type is a struct
		elemType := t.Elem()
		return strings.Contains(elemType.PkgPath(), "birdsfoot")
	}
*/

func convertSlice(slice []interface{}) interface{} {
	if len(slice) == 0 {
		return []string{}
	}

	allStrings := true
	for _, item := range slice {
		if _, ok := item.(string); !ok {
			allStrings = false
			break
		}
	}

	if allStrings {
		result := make([]string, len(slice))
		for i, item := range slice {
			result[i] = item.(string)
		}
		return result
	}

	jsonBytes, _ := json.Marshal(slice)
	return string(jsonBytes)
}

func buildCreateQuery(label string, properties map[string]interface{}) (string, map[string]interface{}) {
	var propParts []string
	params := make(map[string]interface{})

	for key, value := range properties {
		propParts = append(propParts, fmt.Sprintf("%s: $%s", key, key))
		params[key] = value
	}

	cypher := fmt.Sprintf("CREATE (n:%s {%s}) RETURN n.id",
		label, strings.Join(propParts, ", "))

	return cypher, params
}

func buildUpdateQuery(label string, properties map[string]interface{}) (string, map[string]interface{}) {
	var propParts []string
	params := make(map[string]interface{})
	var _idLookup string
	for key, value := range properties {
		switch key {
		case "id":
			_idLookup = fmt.Sprintf("id:'%s'", value)
		case "password":
			if value.(string) != "" {
				propParts = append(propParts, fmt.Sprintf("n.%s = $%s", key, key))
				params[key] = value
			}
		default:
			propParts = append(propParts, fmt.Sprintf("n.%s = $%s", key, key))
			params[key] = value
		}
	}

	cypher := fmt.Sprintf("MERGE (n:%s {%s}) ON MATCH SET %s RETURN n.id",
		label, _idLookup, strings.Join(propParts, ", "))

	return cypher, params
}

func buildQueryCypher(label string, filters map[string]interface{}) (string, map[string]interface{}) {
	cypher := fmt.Sprintf("MATCH (n:%s)", label)
	params := make(map[string]interface{})

	if len(filters) > 0 {
		var whereParts []string
		for key, value := range filters {
			whereParts = append(whereParts, fmt.Sprintf("n.%s = $%s", key, key))
			params[key] = value
		}
		cypher += " WHERE " + strings.Join(whereParts, " AND ")
	}

	cypher += " RETURN n"
	return cypher, params
}

func generateUUID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
