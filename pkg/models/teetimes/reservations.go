package teetimes

import (
	"bigfoot/golf/common/models/account"
	"bigfoot/golf/common/models/db"
	"encoding/json"
	"fmt"
	"time"
)

type Reservation struct {
	ID          string         `json:"id,omitempty"`
	TeeTime     time.Time      `json:"teeTime"`
	BookingUser *account.User  `json:"user,omitempty"`
	Players     []account.User `json:"players"`
	Slot        int64          `json:"slot"`
	Price       float32        `json:"price"`
	SettingType int            `json:"type"`
	Group       string         `json:"group"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
}

type Outing struct {
	ID         string
	Name       string
	StartTime  time.Time
	EndTime    time.Time
	MaxGolfers int64
	TeeTimes   []Reservation
	CreatedAt  time.Time
	UpdatedAt  time.Time
	UpdatedBy  *account.User
}

func (r *Reservation) Save() error {
	_id, err := db.Instance.SaveStruct(r, "Reservation")
	if err != nil {
		fmt.Println(err)
		return err
	}
	r.ID = _id
	var guests []Guest
	for _, guest := range r.Players {
		if guest.ID == "" {
			guests = append(guests, Guest{Name: guest.LastName, Email: guest.Email, Phone: guest.Phone})
		}
		//TODO BUILD ADDING EXISTING USER FUNCTIONALITY
	}
	_rel := db.Relation{NodeN: "User", NodeX: "Reservation", NodeNID: r.BookingUser.ID, NodeXID: r.ID, Name: "BOOKED_TEETIME"}
	if len(guests) > 0 {
		_rel.Property = "guests"
		_g, _ := json.Marshal(guests)
		_rel.Body = string(_g)
	}

	//add Relationship
	err = db.Instance.SaveRelationship(_rel)
	if err != nil {
		return err
	}

	return err
}

func NewReservation(_user *account.User, _players []account.User, _teeTime time.Time, _slot int64, _setting DetailedBlockSettings) Reservation {
	var reserv Reservation
	reserv.BookingUser = _user
	reserv.CreatedAt = time.Now()
	reserv.Players = _players
	reserv.TeeTime = _teeTime
	reserv.Price = _setting.Price
	reserv.SettingType = _setting.Type
	reserv.Group = _setting.Name
	reserv.UpdatedAt = time.Now()
	reserv.Slot = _slot

	return reserv
}

// GetUserReservations retrieves all reservations for a specific user
func GetUserReservations(userID string, includePast bool) ([]Reservation, error) {
	var query string
	if includePast {
		query = fmt.Sprintf(`MATCH (u:User {id: "%s"})-[r:BOOKED_TEETIME]->(res:Reservation)
			WHERE date(res.teeTime) >= date() - duration({months: 12})
			WITH res, r.guests as guests
			RETURN res{.*, guests} as data
			ORDER BY res.teeTime DESC`, userID)
	} else {
		query = fmt.Sprintf(`MATCH (u:User {id: "%s"})-[r:BOOKED_TEETIME]->(res:Reservation)
			WHERE date(res.teeTime) >= date()
			WITH res, r.guests as guests
			RETURN res{.*, guests} as data
			ORDER BY res.teeTime ASC`, userID)
	}

	reservationMaps, err := db.Instance.QueryForMap(query, nil)
	if err != nil {
		return nil, err
	}

	if len(reservationMaps) == 0 {
		return []Reservation{}, nil
	}

	return convertMapsToReservations(reservationMaps), nil
}

// CancelReservation cancels a reservation by marking it as cancelled
func (r *Reservation) Cancel() error {
	query := fmt.Sprintf(`MATCH (res:Reservation {id: "%s"}) 
		SET res.cancelled = true, res.cancelledAt = datetime()
		RETURN res`, r.ID)

	_, err := db.Instance.QueryForMap(query, nil)
	return err
}

// convertMapsToReservations manually converts []map[string]any to []Reservation preserving time.Time locations
func convertMapsToReservations(maps []map[string]any) []Reservation {
	var reservations []Reservation

	for _, m := range maps {
		var reservation Reservation

		if id, ok := m["id"].(string); ok {
			reservation.ID = id
		}
		if teeTime, ok := m["teeTime"].(time.Time); ok {
			reservation.TeeTime = teeTime
		}
		if slot, ok := m["slot"].(int64); ok {
			reservation.Slot = slot
		}
		if price, ok := m["price"].(float64); ok {
			reservation.Price = float32(price)
		}
		if settingType, ok := m["type"].(int64); ok {
			reservation.SettingType = int(settingType)
		}
		if group, ok := m["group"].(string); ok {
			reservation.Group = group
		}
		if createdAt, ok := m["createdAt"].(time.Time); ok {
			reservation.CreatedAt = createdAt
		}
		if updatedAt, ok := m["updatedAt"].(time.Time); ok {
			reservation.UpdatedAt = updatedAt
		}

		// Handle guests from relationship property
		if guestsStr, ok := m["guests"].(string); ok && guestsStr != "" {
			var guests []Guest
			if err := json.Unmarshal([]byte(guestsStr), &guests); err == nil {
				for _, guest := range guests {
					reservation.Players = append(reservation.Players, account.User{
						LastName: guest.Name,
						Email:    guest.Email,
						Phone:    guest.Phone,
					})
				}
			}
		}

		reservations = append(reservations, reservation)
	}

	return reservations
}
