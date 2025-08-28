package handlers

import (
	"birdsfoot/app/models"
	"encoding/json"
	"net/http"
)

// Helper function to send JSON responses
func sendJSONResponse(w http.ResponseWriter, statusCode int, response models.Response) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	response := models.Response{
		Success: false,
		Error:   "Endpoint not found",
	}
	w.Header().Set("Content-Type", "application/json")
	sendJSONResponse(w, http.StatusNotFound, response)
}
