package transactions

import (
	"birdsfoot/app/handlers/sessionmgr"
	"birdsfoot/app/models/account"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/smtp"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// PUT /api/chat/{id} - Get user by ID
func SaveUserHandler(w http.ResponseWriter, r *http.Request) {

	var user account.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := user.Save()
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)

}

func SendEmailCodeHandler(w http.ResponseWriter, r *http.Request) {

	var user account.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	from := os.Getenv("GMAIL_USER")
	password := os.Getenv("GMAIL_PASS")
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	verifyCode := randomSixDigit()
	message := fmt.Sprintf("Subject: Verify Account\n\nPlease enter this code into your app to verify your account:\n %s", verifyCode)

	//store code in session
	_expiresIn := time.Now().Add(time.Minute * 10)
	err := sessionmgr.SeshStore.StoreString(sessionmgr.SeshData{Code: verifyCode, ID: user.ID, ExpiresAt: _expiresIn}, r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	auth := smtp.PlainAuth("", from, password, smtpHost)
	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{user.Email}, []byte(message))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	//json.NewEncoder(w).Encode(user)

}

func VerifyCodeHandler(w http.ResponseWriter, r *http.Request) {
	var _user account.User
	if err := json.NewDecoder(r.Body).Decode(&_user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_data := sessionmgr.SeshStore.GetString(r, w)
	if !_data.Ok {
		http.Error(w, _data.Message, http.StatusConflict)
		return
	}
	// Check if code has expired
	if time.Now().After(_data.ExpiresAt) {
		// Clear expired session
		http.Error(w, _data.Message, http.StatusConflict)
		return
	}

	// Verify user ID and code
	if _data.ID != _user.ID && _user.TempStr != _data.Code {
		http.Error(w, _data.Message, http.StatusConflict)
		return
	}
	//user has verified email, check that
	_user.IsVerified = true
	_user.Save()

	w.Header().Set("Content-Type", "application/json")
	//send success message
	response := make(map[string]interface{})
	response["success"] = true
	response["message"] = "Verification successful"

	json.NewEncoder(w).Encode(response)
}

func randomSixDigit() string {
	return fmt.Sprintf("%06d", rand.IntN(1000000))
}

func UpdatePW(w http.ResponseWriter, r *http.Request) {

	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Hash password
	_newPW, ok := input["password"].(string)
	if !ok {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(_newPW), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}
	var user account.User
	user.ID = input["id"].(string)
	user.Password = string(hashedPassword)

	err = user.UpdatePW()
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)

}
