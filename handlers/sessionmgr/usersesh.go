package sessionmgr

import (
	"encoding/gob"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/sessions"
)

type BirdSessionMgr struct {
	cookieStr *sessions.CookieStore
}

type SeshData struct {
	ID        string    `json:"id"`
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expires"`
	Message   string    `json:"message"`
	Ok        bool      `json:"ok"`
}

var (
	// Session store with a secret key (use a strong random key in production)
	SeshStore BirdSessionMgr
)

func NewSessionMgr() {
	gob.Register(SeshData{})
	seshkey := os.Getenv("SESSION_KEY")
	SeshStore.cookieStr = sessions.NewCookieStore([]byte(seshkey))
}

func (s *BirdSessionMgr) StoreString(_data SeshData, r *http.Request, w http.ResponseWriter) error {
	// Get session
	session, err := SeshStore.cookieStr.Get(r, "bird-session")
	if err != nil {
		return err
	}

	// Store in session
	session.Values["sesh_data"] = _data
	session.Options.MaxAge = 600 // 10 minutes in seconds

	// Save session
	if err := session.Save(r, w); err != nil {
		return err
	}
	return nil
}

func (s *BirdSessionMgr) GetString(r *http.Request, w http.ResponseWriter) SeshData {
	// Get session
	session, err := SeshStore.cookieStr.Get(r, "bird-session")
	if err != nil {
		return SeshData{Ok: false, Message: "Failed to Get Session"}
	}

	// Check if code data exists in session
	codeDataInterface, exists := session.Values["sesh_data"]
	if !exists {
		return SeshData{Ok: false, Message: "No verification code found"}
	}

	codeData, ok := codeDataInterface.(SeshData)
	if !ok {
		return SeshData{Ok: false, Message: "Failed to Get Code"}

	}
	codeData.Ok = true
	return codeData
}
