package account

import (
	"birdsfoot/app/models/db"
	"fmt"
	"log"
	"time"

	"github.com/mitchellh/mapstructure"
)

type User struct {
	ID        string    `json:"id,omitempty"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Phone     string    `json:"phone"` // Foreign key relationship
	DOB       string    `json:"dob"`   // Self-referential relationship
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	//auth fields
	Password   string `json:"password"`
	Provider   string `json:"provider"` // "local", "google", "apple"
	ProviderID string `json:"provider_id"`
	Avatar     string `json:"avatar"`
	IsVerified bool   `json:"is_verified"`
	IsAdmin    bool   `json:"is_admin"`
	TempStr    string
}

func (u *User) Save() error {

	if u.Email == "" {
		return fmt.Errorf("no email supplied")
	}
	u.UpdatedAt = time.Now()
	userID, err := db.Instance.SaveStruct(u, "User")
	if err != nil {
		log.Printf("Error saving user with relationships: %v", err)
		return err
	}
	u.ID = userID
	fmt.Printf("Saved user with relationships, ID: %s\n", userID)
	return nil
}

func (u *User) UpdatePW() error {

	var node db.DynamicNode
	if u.ID == "" || u.Password == "" {
		return fmt.Errorf("invalid identifier/password")
	}

	node.Label = "User"
	node.Properties = make(map[string]interface{})
	node.Properties["id"] = u.ID
	node.Properties["password"] = u.Password

	_, err := db.Instance.SaveDynamicNode(node)
	return err
}

type Company struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name"`
	Industry string `json:"industry"`
	Location string `json:"location"`
}

func QueryUsers(query map[string]interface{}) ([]User, error) {
	return queryUsers(query)
}
func QueryUser(query map[string]interface{}) (*User, error) {
	users, err := queryUsers(query)
	if len(users) > 0 {
		return &users[0], err
	}
	return nil, err
}
func queryUsers(query map[string]interface{}) ([]User, error) {
	users, err := db.Instance.QueryNodes("User", query) // depth of 2
	if err != nil {
		return nil, err
	}
	var _users []User
	if len(users) > 0 {
		config := &mapstructure.DecoderConfig{
			Result:  &_users,
			TagName: "json", // Use JSON tags instead of mapstructure tags
		}

		decoder, err := mapstructure.NewDecoder(config)
		if err != nil {
			return nil, err
		}

		err = decoder.Decode(users)
		if err != nil {
			return nil, err
		}
	}
	return _users, nil
}
