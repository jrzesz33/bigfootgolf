package handlers

import (
	"bigfoot/golf/common/models"
	"bigfoot/golf/common/models/auth"
	"encoding/json"
	"net/http"
)

// POST /api/chat/bedrock - Handle chat request with AWS Bedrock
func GetChatHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var message models.AgentChatRequest
	if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
		response := models.Response{
			Success: false,
			Error:   "Invalid JSON in chat request",
		}
		sendJSONResponse(w, http.StatusBadRequest, response)
		return
	}

	chatAgent := models.NewGolfAgent()

	// Get user ID from header (set by authentication middleware)
	userID, _ := r.Context().Value(auth.USERIDKEY).(string)
	if userID == "" {
		//error no user id so return
		sendJSONResponse(w, http.StatusInternalServerError, models.Response{Success: false, Error: "no user id found"})
		return
	}

	chatResponse, err := chatAgent.SendWithMessage(userID, message)

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
