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

	// set start date
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

func BookTime(w http.ResponseWriter, r *http.Request) {
	var input teetimes.Reservation
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if input.BookingUser == nil || len(input.Players) < 1 {
		http.Error(w, "No User Found", http.StatusForbidden)
		return
	}

	err := input.Save()
	if err != nil {
		http.Error(w, "Error with Transaction, Try Again Later", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(input)
}

// GetUserReservations retrieves all reservations for the authenticated user
func GetUserReservations(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID") // Assuming user ID is passed in header after auth
	if userID == "" {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	var input map[string]bool
	includePast := false
	if r.Method == "POST" {
		if err := json.NewDecoder(r.Body).Decode(&input); err == nil {
			if val, exists := input["includePast"]; exists {
				includePast = val
			}
		}
	} else {
		includePast = r.URL.Query().Get("includePast") == "true"
	}

	reservations, err := teetimes.GetUserReservations(userID, includePast)
	if err != nil {
		http.Error(w, "Error retrieving reservations", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reservations)
}

// CancelReservation cancels a specific reservation
func CancelReservation(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	var input map[string]string
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	reservationID, exists := input["reservationId"]
	if !exists || reservationID == "" {
		http.Error(w, "Reservation ID required", http.StatusBadRequest)
		return
	}

	// First verify the reservation belongs to the user
	reservations, err := teetimes.GetUserReservations(userID, false)
	if err != nil {
		http.Error(w, "Error verifying reservation ownership", http.StatusInternalServerError)
		return
	}

	var targetReservation *teetimes.Reservation
	for i := range reservations {
		if reservations[i].ID == reservationID {
			targetReservation = &reservations[i]
			break
		}
	}

	if targetReservation == nil {
		http.Error(w, "Reservation not found or not owned by user", http.StatusNotFound)
		return
	}

	err = targetReservation.Cancel()
	if err != nil {
		http.Error(w, "Error cancelling reservation", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "cancelled"})
}
