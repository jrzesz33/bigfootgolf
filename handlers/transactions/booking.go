package transactions

import (
	"birdsfoot/app/models/teetimes"
	"encoding/json"
	"net/http"
	"time"
)

func GetTeeTimes(w http.ResponseWriter, r *http.Request) {
	var input map[string]time.Time
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Hash password
	_start := input["start"]

	var booking teetimes.BookingEngine

	_days, err := booking.GetDayTeeTimes(_start)
	if err != nil {
		http.Error(w, "Issue with Search", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(_days)
}
