package teetimes

import (
	"birdsfoot/app/models/account"
	"birdsfoot/app/models/db"
	"fmt"
	"time"
)

type ReservationBlock struct {
	ID           string                    `json:"id,omitempty"`
	Season       string                    `json:"season"`
	BeginDate    time.Time                 `json:"beginDate"`
	EndDate      time.Time                 `json:"endDate"`
	FirstTeeTime time.Duration             `json:"firstTeeTime"`
	LastTeeTime  time.Duration             `json:"lastTeeTime"`
	Gap          time.Duration             `json:"gap"`
	Dates        map[time.Time]ReservedDay `json:"dates"`
	BlockDetails []DetailedBlockSettings   `json:"customBlocks"`
	Outings      []Outing
	CreatedAt    time.Time     `json:"createdAt"`
	UpdatedAt    time.Time     `json:"updatedAt"`
	UpdatedBy    *account.User `json:"updatedBy"`
}

func NewReservationBlock(season Season) ReservationBlock {
	var block ReservationBlock
	block.BeginDate = season.BeginDate
	block.EndDate = season.EndDate
	block.CreatedAt = time.Now()
	block.UpdatedAt = time.Now()
	block.BlockDetails = GetDetailedBlockSettings(season)

	return block
}

func (r *ReservationBlock) Save() error {

	_strOut, err := db.Instance.SaveStruct(r, "ReservationBlock")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(_strOut)
	return err
}
