package teetimes

import (
	"birdsfoot/app/helper"
	"birdsfoot/app/models/db"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type BookingEngine struct {
	ActiveBlock  ReservationBlock
	ActiveBlocks []ReservationBlock
}

type AppConfigFile struct {
	CreatedAt time.Time `yaml:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `yaml:"updatedAt" json:"updatedAt"`
	Seasons   []Season  `yaml:"seasons" json:"seasons"`
}

type SettingType int

const (
	Weekday_Morning SettingType = iota
	Weekday_Midday
	Weekday_Afternoon
	Weekend_Morning
	Weekend_Midday
	Weekend_Afternoon
	Holiday
	Daily_Deal
)

type DetailedBlockSettings struct {
	ID            string    `yaml:"id" json:"id"`
	Name          string    `yaml:"name" json:"name"`
	Type          int       `yaml:"type" json:"type"`
	BeginOverride time.Time `yaml:"beginOverride" json:"beginOverride"`
	EndOverride   time.Time `yaml:"endOverride" json:"endOverride"`
	Price         float32   `yaml:"price" json:"price"`
	IsAvail       bool      `yaml:"isAvail" json:"isAvail"`
}

func GetDetailedBlockSettings(season Season) []DetailedBlockSettings {

	var dbs []DetailedBlockSettings

	//break out the holiday, weekend, morning, afternoon, and evening rates

	return dbs
}

func (b *BookingEngine) BookSlot(reservation Reservation) error {

	_day := helper.TruncateToDay(reservation.TeeTime)
	b.ActiveBlock.Dates[_day].Reservations[int(reservation.Slot)] = reservation

	return nil
}

func (b *BookingEngine) GetDayTeeTimes(_date time.Time) ([]ReservedDay, error) {

	//get the tee times
	query := fmt.Sprintf(`MATCH (n:ReservedDay) WHERE date(n.day) = date("%s")
		OPTIONAL MATCH (n)-[r*1..1]-(related)
		WITH n, collect(DISTINCT related{.*}) as t
		RETURN { id: n.id, day: n.day, times: t } as data`, _date.Format(time.DateOnly))

	dayWithRelationships, err := db.Instance.QueryForJSON(query, nil) // depth of 2
	if err != nil {
		log.Printf("Error querying with relationships: %v", err)
		return nil, err
	}
	var days []ReservedDay
	err = json.Unmarshal(dayWithRelationships, &days)
	if err != nil {
		return nil, err
	}

	if len(days) == 0 {
		//no block so create times
	}

	return days, nil

}

func (b *BookingEngine) AddSeason(season string, dts []time.Time) {
	//differentiate weekday, holiday, morning Afternoon Times

}

func (d *DetailedBlockSettings) Save() (string, error) {
	//differentiate weekday, holiday, morning Afternoon Times
	_id, err := db.Instance.SaveStruct(d, "DetailedBlockSettings")
	if err != nil || _id == "" {
		fmt.Println(err)
		return "", err
	}
	d.ID = _id
	return _id, nil
}
