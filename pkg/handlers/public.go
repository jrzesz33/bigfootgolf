package handlers

import (
	"bigfoot/golf/common/handlers/transactions"
	"bigfoot/golf/common/models/courses"
	"bigfoot/golf/common/models/weather"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func RegisterPublicRoutes(router *mux.Router) {

	// Create weather handler with 15-minute cache
	weatherHandler := weather.NewWeatherHandler(
		"https://api.weather.gov/gridpoints/PBZ/87,81/forecast",
		15*time.Minute,
	)

	// Public routes
	router.HandleFunc("/weather", weatherHandler.ServeHTTP).Methods("GET")
	router.HandleFunc("/teetimes", transactions.GetTeeTimes).Methods("POST")
	router.HandleFunc("/courses", GetActiveCourses).Methods("GET")
	router.HandleFunc("/realtime", transactions.GetRealtimeTeeTimes).Methods("POST")
}

// GetActiveCourses returns all active courses
func GetActiveCourses(w http.ResponseWriter, r *http.Request) {
	allCourses, err := courses.GetAll()
	if err != nil {
		http.Error(w, "Error retrieving courses", http.StatusInternalServerError)
		return
	}

	// Filter only active courses
	var activeCourses []courses.Course
	for _, course := range allCourses {
		if course.Active {
			activeCourses = append(activeCourses, course)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(activeCourses)
}
