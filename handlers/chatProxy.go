package handlers

import (
	"birdsfoot/app/controllers"
	"birdsfoot/app/models"
	"birdsfoot/app/models/anthropic"
	"encoding/json"
	"net/http"
)

// PUT /api/chat/{id} - Get user by ID
func GetChatHandler(w http.ResponseWriter, r *http.Request) {

	var message anthropic.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
		response := models.Response{
			Success: false,
			Error:   "Errorz with calling Chat Request",
		}
		sendJSONResponse(w, http.StatusNotFound, response)
	}
	_claudeClient := controllers.NewAgentController()
	_claudeClient.HandleChat(message)

}
