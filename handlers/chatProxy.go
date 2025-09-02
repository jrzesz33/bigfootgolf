package handlers

import (
	"birdsfoot/app/controllers"
	"birdsfoot/app/models"
	"birdsfoot/app/models/anthropic"
	"encoding/json"
	"net/http"
)

// POST /api/chat - Handle chat request with Claude
func GetChatHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var message anthropic.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
		response := models.Response{
			Success: false,
			Error:   "Invalid JSON in chat request",
		}
		sendJSONResponse(w, http.StatusBadRequest, response)
		return
	}

	_claudeClient := controllers.NewAgentController()
	
	// Get user ID from header (set by authentication middleware)
	userID := r.Header.Get("X-User-ID")
	if userID != "" {
		_claudeClient.SetUserID(userID)
	}
	
	chatResponse, err := _claudeClient.HandleChat(message)
	
	if err != nil {
		response := models.Response{
			Success: false,
			Error:   "Error processing chat request: " + err.Error(),
		}
		sendJSONResponse(w, http.StatusInternalServerError, response)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(chatResponse)
}
