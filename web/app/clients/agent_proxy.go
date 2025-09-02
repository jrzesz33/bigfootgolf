package clients

import (
	"bigfoot/golf/common/models/anthropic"
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
	resp, err := SendPostWithAuth(url, string(message))
	if err.Code != 200 {
		fmt.Println("Error with Proxy ", err)
		return nil, err.BError
	}
	var chatResp anthropic.ChatResponse
	erx := json.Unmarshal(resp, &chatResp)
	if erx != nil {
		return nil, erx
	}
	return &chatResp, nil
}
