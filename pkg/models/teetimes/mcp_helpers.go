package teetimes

import (
	"bigfoot/golf/common/models/db"
	"context"
	"fmt"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// GetUserReservationsForMCP wraps the existing GetUserReservations for MCP usage
func GetUserReservationsForMCP(userID string) ([]Reservation, error) {
	// Use the existing GetUserReservations function with includePast=true to get all reservations
	return GetUserReservations(userID, true)
}

// CreateReservation creates a new reservation for a user
func CreateReservation(userID string, teeTime time.Time, players int) (*Reservation, error) {
	ctx := context.Background()
	driver := db.Instance

	session := driver.NewWriteSession(ctx)
	defer session.Close(ctx)

	reservationID := generateReservationID()

	query := `
		MATCH (u:User {id: $userID})
		MATCH (b:ReservationBlock {date: $date, time: $time, available: true})
		WHERE b.slots >= $players
		CREATE (r:Reservation {
			id: $reservationID,
			date: $date,
			time: $time,
			players: $players,
			status: 'confirmed',
			created: datetime()
		})
		CREATE (u)-[:HAS_RESERVATION]->(r)
		CREATE (r)-[:BOOKS]->(b)
		SET b.slots = b.slots - $players
		RETURN r
	`

	result, err := session.Run(ctx, query, map[string]interface{}{
		"userID":        userID,
		"reservationID": reservationID,
		"date":          teeTime.Format("2006-01-02"),
		"time":          teeTime.Format("15:04"),
		"players":       players,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create reservation: %v", err)
	}

	if result.Next(ctx) {
		record := result.Record()
		resNode, _ := record.Get("r")
		if resNode != nil {
			res := nodeToReservation(resNode)
			return &res, nil
		}
	}

	return nil, fmt.Errorf("failed to create reservation")
}

// CancelReservation cancels an existing reservation
func CancelReservation(userID, reservationID string) error {
	ctx := context.Background()
	driver := db.Instance

	session := driver.NewWriteSession(ctx)
	defer session.Close(ctx)

	query := `
		MATCH (u:User {id: $userID})-[:HAS_RESERVATION]->(r:Reservation {id: $reservationID})
		MATCH (r)-[:BOOKS]->(b:ReservationBlock)
		SET r.status = 'cancelled'
		SET b.slots = b.slots + r.players
		RETURN r
	`

	result, err := session.Run(ctx, query, map[string]interface{}{
		"userID":        userID,
		"reservationID": reservationID,
	})
	if err != nil {
		return fmt.Errorf("failed to cancel reservation: %v", err)
	}

	if !result.Next(ctx) {
		return fmt.Errorf("reservation not found or unauthorized")
	}

	return nil
}

// GetAvailableTeeTimes returns available tee times for a given date
func GetAvailableTeeTimes(date time.Time, timeRange string, players int) ([]ReservationBlock, error) {
	ctx := context.Background()
	driver := db.Instance.Driver

	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	dateStr := date.Format("2006-01-02")

	var timeFilter string
	switch timeRange {
	case "morning":
		timeFilter = "AND b.time >= '06:00' AND b.time < '12:00'"
	case "midday":
		timeFilter = "AND b.time >= '11:00' AND b.time < '14:00'"
	case "afternoon":
		timeFilter = "AND b.time >= '14:00' AND b.time < '18:00'"
	default:
		timeFilter = ""
	}

	query := fmt.Sprintf(`
		MATCH (b:ReservationBlock)
		WHERE b.date = $date 
		AND b.available = true 
		AND b.slots >= $players
		%s
		RETURN b
		ORDER BY b.time
	`, timeFilter)

	result, err := session.Run(ctx, query, map[string]interface{}{
		"date":    dateStr,
		"players": players,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query available tee times: %v", err)
	}

	var blocks []ReservationBlock
	for result.Next(ctx) {
		record := result.Record()
		blockNode, _ := record.Get("b")
		if blockNode != nil {
			block := nodeToReservationBlock(blockNode)
			blocks = append(blocks, block)
		}
	}

	return blocks, nil
}

// GetCourseConditions returns current course conditions
func GetCourseConditions() (map[string]interface{}, error) {
	// This would typically fetch from a database or external service
	// For now, returning mock data
	return map[string]interface{}{
		"greens":     "Firm and Fast",
		"fairways":   "Good",
		"cartPath":   "Open",
		"range":      "Open",
		"notes":      "Course in excellent condition",
		"lastUpdate": time.Now().Format(time.RFC3339),
	}, nil
}

// Helper function to generate reservation ID
func generateReservationID() string {
	return fmt.Sprintf("RES-%d", time.Now().Unix())
}

// Helper functions to convert Neo4j nodes to structs
func nodeToReservation(node interface{}) Reservation {
	// Implementation would convert Neo4j node to Reservation struct
	// This is a placeholder
	return Reservation{}
}

func nodeToSeason(node interface{}) *Season {
	// Implementation would convert Neo4j node to Season struct
	// This is a placeholder
	return &Season{}
}

func nodeToReservationBlock(node interface{}) ReservationBlock {
	// Implementation would convert Neo4j node to ReservationBlock struct
	// This is a placeholder
	return ReservationBlock{}
}
