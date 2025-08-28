package teetimes

import (
	"birdsfoot/app/models/account"
	"birdsfoot/app/models/db"
	"fmt"
	"time"
)

type Reservation struct {
	ID          string         `json:"id,omitempty"`
	TeeTime     time.Time      `json:"teeTime"`
	BookingUser *account.User  `json:"user,omitempty"`
	Players     []account.User `json:"players"`
	Slot        int64          `json:"slot"`
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
	_strOut, err := db.Instance.SaveStruct(r, "Reservation")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(_strOut)
	return err
}

func NewReservation(_user *account.User, _players []account.User, _teeTime time.Time, _slot int64) Reservation {
	var reserv Reservation
	reserv.BookingUser = _user
	reserv.CreatedAt = time.Now()
	reserv.Players = _players
	reserv.TeeTime = _teeTime
	reserv.UpdatedAt = time.Now()
	reserv.Slot = _slot

	return reserv
}
