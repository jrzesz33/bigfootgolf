package helper

import (
	"encoding/json"
	"fmt"
	"time"
)

const DATE_LAYOUT string = "2006-01-02T15:04:05Z"

func TruncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// Method 1: Using JSON Marshal/Unmarshal (Most Common)
func MapToStructJSON(data map[string]any, result interface{}) error {
	// Convert map to JSON bytes
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshaling map to JSON: %v", err)
	}

	// Convert JSON bytes to struct
	err = json.Unmarshal(jsonBytes, result)
	if err != nil {
		return fmt.Errorf("error unmarshaling JSON to struct: %v", err)
	}

	return nil
}

// Method 1: Using JSON Marshal/Unmarshal (Most Common)
func MapsToStructJSON(data []map[string]any, result any) (any, error) {
	// Convert map to JSON bytes
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("error marshaling map to JSON: %v", err)
	}

	// Convert JSON bytes to struct
	err = json.Unmarshal(jsonBytes, &result)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON to struct: %v", err)
	}

	return result, nil
}

func GetSeasonsMetereological(year int) map[string][]time.Time {
	seasons := make(map[string][]time.Time)

	// Meteorological seasons:
	// Spring: March 1 - May 31
	// Summer: June 1 - August 31
	// Fall: September 1 - November 30
	// Winter: December 1 - February 28/29 (next year)

	springBegin := time.Date(year, time.March, 1, 0, 0, 0, 0, time.Local)
	springEnd := time.Date(year, time.May, 31, 23, 59, 59, 0, time.Local)

	summerBegin := time.Date(year, time.June, 1, 0, 0, 0, 0, time.Local)
	summerEnd := time.Date(year, time.August, 31, 23, 59, 59, 0, time.Local)

	fallBegin := time.Date(year, time.September, 1, 0, 0, 0, 0, time.Local)
	fallEnd := time.Date(year, time.November, 30, 23, 59, 59, 0, time.Local)

	winterBegin := time.Date(year, time.December, 1, 0, 0, 0, 0, time.Local)

	// Handle leap year for February
	var winterEndDay int
	if isLeapYear(year + 1) {
		winterEndDay = 29
	} else {
		winterEndDay = 28
	}
	winterEnd := time.Date(year+1, time.February, winterEndDay, 23, 59, 59, 0, time.Local)

	seasons["spring"] = []time.Time{springBegin, springEnd}
	seasons["summer"] = []time.Time{summerBegin, summerEnd}
	seasons["fall"] = []time.Time{fallBegin, fallEnd}
	//seasons["autumn"] = []time.Time{fallBegin, fallEnd}
	seasons["winter"] = []time.Time{winterBegin, winterEnd}

	return seasons
}

// Helper function to check if a year is a leap year
func isLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}
