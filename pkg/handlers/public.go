package handlers

import (
	"bigfoot/golf/common/handlers/transactions"
	"bigfoot/golf/common/models/weather"
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
}
