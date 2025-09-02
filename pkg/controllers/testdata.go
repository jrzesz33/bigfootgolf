package controllers

import (
	"bigfoot/golf/common/models/account"
	"bigfoot/golf/common/models/db"
	"bigfoot/golf/common/models/teetimes"

	"log"
	"time"
)

// Example usage demonstrating relationship mapping
func SetupDevEnvironment() {

	var eng teetimes.BookingEngine
	//firstTee := time.Hour*6 + time.Minute*30
	//lastTee := time.Hour*19 + time.Minute*50
	if isTestDataSetup() {
		return
	}
	/*	//check if there are tee times today
		_today := time.Now().Truncate(24 * time.Hour).Local()
		_times, _ := eng.GetDayTeeTimes(_today)
		if len(_times) < 1 {
			eng.ActiveBlock = teetimes.NewReservationBlock(_today, _today.Add(time.Hour*24), 10*time.Minute, firstTee, lastTee)
			eng.ActiveBlock.Save()
		}
		return

	}*/
	// Example 1: Save struct with relationship definitions
	user := account.User{
		Email:     "john@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Phone:     "123-123-1234",
		DOB:       "01/01/1990",
		CreatedAt: time.Now(),
	}

	//user.Save()

	//start := time.Date(2025, time.June, 1, 0, 0, 0, 0, time.Now().Local().Location())
	//end := time.Date(2025, time.August, 1, 0, 0, 0, 0, time.Now().Local().Location())

	//eng.ActiveBlock = teetimes.NewReservationBlock(start, end, 10*time.Minute, firstTee, lastTee)

	//eng.ActiveBlock.Save()

	//add the reservations to the block
	testDay := time.Date(2025, time.July, 4, 0, 0, 0, 0, time.Now().Local().Location())
	slot := 4
	_booking := eng.ActiveBlock.Dates[testDay].Times[slot]
	_booking.BookingUser = &user
	_booking.Players = append(_booking.Players, user)

	eng.BookSlot(_booking)

	/* Example 4: Query with relationships
	dayWithRelationships, err := db.Instance.QueryNodesWithRelationships("ReservedDay", map[string]interface{}{
		"day": "2025-07-22T00:00:00Z",
	}, 2) // depth of 2

	if err != nil {
		log.Printf("Error querying with relationships: %v", err)
	} else {
		for _, day := range dayWithRelationships {
			fmt.Printf("User: %v\n", day.Node)
			fmt.Printf("Relationships: %v\n", day.Relationships)
			fmt.Printf("Related nodes: %v\n", day.RelatedNodes)
		}
	}*/
}

func isTestDataSetup() bool {
	layout := "2006-01-02T15:04:05Z"
	// Format the time object into a string
	formattedTime := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local).Format(layout)
	// Example 4: Query with relationships
	dayWithRelationships, err := db.Instance.QueryNodes("ReservedDay", map[string]interface{}{
		"day": formattedTime,
	}) // depth of 2

	if err != nil {
		log.Printf("Error querying with relationships: %v", err)
	} else {
		/*for _, day := range dayWithRelationships {
			fmt.Printf("Day: %v\n", day.Node)
			fmt.Printf("Relationships: %v\n", day.Relationships)
			fmt.Printf("Related nodes: %v\n", day.RelatedNodes)
		}*/
		if len(dayWithRelationships) > 0 {
			return true
		}
	}

	return true
}
