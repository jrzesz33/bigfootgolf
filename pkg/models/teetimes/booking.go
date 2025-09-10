package teetimes

import (
	"bigfoot/golf/common/models/db"
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

func BookTeeTime(res Reservation) error {
	if res.BookingUser == nil {
		return fmt.Errorf("no user found")
	}
	if len(res.Players) == 0 {
		res.Players = append(res.Players, *res.BookingUser)
	}
	return res.Save()
}

func (b *BookingEngine) GetDayTeeTimes(_date time.Time) ([]ReservedDay, error) {
	var days []Reservation
	//slotQuery := ""
	//if slot != nil && *slot >= 0 {
	//	slotQuery = fmt.Sprintf(` {slot: %d}`, slot)
	//}
	//get the tee times
	query := fmt.Sprintf(`MATCH (n:Reservation) WHERE date(n.createdAt) = date("%s")
		MATCH (u:User)-[r:BOOKED_TEETIME]->(n)
		WITH n, u {.*} as user, COLLECT(u {.*}) as players
		RETURN n{.* , user, players} as data`, _date.Format(time.DateOnly))

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
	var daysOut []ReservedDay
	//no block so create times
	_seas, err := GetSeason(_date)
	if err != nil {
		return nil, err
	}
	if _seas != nil {
		_newDay := NewReservedDay(_date, *_seas, days)
		daysOut = append(daysOut, _newDay)
	}

	return daysOut, nil

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
