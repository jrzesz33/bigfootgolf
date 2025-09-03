package controllers

import (
	"bigfoot/golf/common/models/db"
	"fmt"

	"log"
	"time"
)

// Example usage demonstrating relationship mapping
func SetupDevEnvironment() {

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
	fmt.Println("test")
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
