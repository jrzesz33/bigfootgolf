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

// SettingType constants define different tee time settings
const (
	WeekdayMorning SettingType = iota
	WeekdayMidday
	WeekdayAfternoon
	WeekendMorning
	WeekendMidday
	WeekendAfternoon
	Holiday
	DailyDeal
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
	for i, res := range b.ActiveBlock.Dates[_day].Times {
		if res.Slot == reservation.Slot {
			b.ActiveBlock.Dates[_day].Times[i] = reservation
			return nil
		}
	}
	return nil
}

func (b *BookingEngine) GetDayTeeTimes(_date time.Time) ([]ReservedDay, error) {
	var days []ReservedDay

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
	if dayWithRelationships != nil {
		err = json.Unmarshal(dayWithRelationships, &days)
		if err != nil {
			return nil, err
		}
	}

	if len(days) == 0 {
		//no block so create times
		_seas, err := GetSeason(_date)
		if err != nil {
			return nil, err
		}
		if _seas != nil {
			_newDay := NewReservedDay(_date, *_seas)
			days = append(days, _newDay)
		}
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
func (d *DetailedBlockSettings) MatchesType(_day time.Time, _time time.Time) bool {
	dayType := []int{int(WeekdayMorning), int(WeekdayAfternoon)}
	if _day.Weekday() == time.Saturday || _day.Weekday() == time.Sunday {
		dayType = []int{int(WeekdayMorning), int(WeekendAfternoon)}
	}
	if d.Type >= dayType[0] && d.Type <= dayType[1] {
		return true
	}
	return false
}

func mapToBlockDetails(settingMap map[string]interface{}) DetailedBlockSettings {
	var dbs DetailedBlockSettings

	if id, ok := settingMap["id"].(string); ok {
		dbs.ID = id
	}
	if name, ok := settingMap["name"].(string); ok {
		dbs.Name = name
	}
	if settingType, ok := settingMap["type"].(int64); ok {
		dbs.Type = int(settingType)
	}
	if beginOverride, ok := settingMap["beginOverride"].(time.Time); ok {
		dbs.BeginOverride = beginOverride
	}
	if endOverride, ok := settingMap["endOverride"].(time.Time); ok {
		dbs.EndOverride = endOverride
	}
	if price, ok := settingMap["price"].(float64); ok {
		dbs.Price = float32(price)
	}
	if isAvail, ok := settingMap["isAvail"].(bool); ok {
		dbs.IsAvail = isAvail
	}
	return dbs
}
