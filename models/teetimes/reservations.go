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
	_strOut, err := db.Instance.SaveStruct(r, "Reservation")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(_strOut)
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
