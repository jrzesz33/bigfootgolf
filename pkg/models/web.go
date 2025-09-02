package models

import "strings"

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type BError struct {
	Request  string `json:"request"`
	Code     int    `json:"code"`
	BError   error  `json:"error"`
	Redirect string `json:"redirect"`
}

func (b *BError) FriendlyMsg() string {
	strOut := "There was a problem with your request, please try again later."

	switch b.Code {
	case 401:
		strOut = "You do not appear to be logged in, please login."
		b.Redirect = "login"
	case 404:
		strOut = "We need a mulligan. This is embarassing, this page or content does not exist."
	case 409:
		if strings.EqualFold(b.Request, "./auth/register") || strings.EqualFold(b.Request, "./api/userupdate") {
			strOut = "The Email address is already in use, please select Login or Forgot Password."
		}

	}

	return strOut
}
