package clients

import (
	"birdsfoot/app/models/anthropic"
	"encoding/json"
	"fmt"
)

// CallAgentProxy sends a chat request to the agent proxy API endpoint
func CallAgentProxy(request anthropic.ChatRequest) (*anthropic.ChatResponse, error) {
	url := "./api/chat"
	message, erz := json.Marshal(request)
	if erz != nil {
		return nil, erz
	}
	resp, err := SendPostWithPayload(url, string(message))
	if err != nil {
		fmt.Println("Error with Proxy ", err)
		return nil, err
	}
	var chatResp anthropic.ChatResponse
	erx := json.Unmarshal(resp, &chatResp)
	if erx != nil {
		return nil, err
	}
	return &chatResp, nil
}
