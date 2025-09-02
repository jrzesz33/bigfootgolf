package teetimes

import (
	"birdsfoot/app/helper"
	"birdsfoot/app/models/db"
	"birdsfoot/app/models/weather"
	"fmt"
	"log"
	"time"
)

type Season struct {
	ID              string                      `yaml:"id" json:"id"`
	Year            int                         `yaml:"year" json:"year"`
	Name            string                      `yaml:"name" json:"name"`
	BeginDate       time.Time                   `yaml:"beginDate" json:"beginDate"`
	EndDate         time.Time                   `yaml:"endDate" json:"endDate"`
	SolarTimes      *weather.ParsedSolarResults `yaml:"solarTimes" json:"solarTimes"`
	FirstTeeTime    time.Time                   `yaml:"firstTeeTime" json:"firstTeeTime"`
	LastTeeTime     time.Time                   `yaml:"lastTeeTime" json:"lastTeeTime"`
	Gap             time.Duration               `yaml:"gap" json:"gap"`
	IsOpen          bool                        `yaml:"isOpen" json:"isOpen"`
	DefaultSettings []DetailedBlockSettings     `yaml:"defaultSettings" json:"defaultSettings"`
	OverideSettings []DetailedBlockSettings     `yaml:"overideSettings" json:"overideSettings"`
}

func NewSeason(year int, name string, begin time.Time, end time.Time) Season {
	var lat, lon float32
	var err error
	lat = 40.745152
	lon = -79.665367

	var seas Season
	seas.Year = year
	seas.BeginDate = begin
	seas.Name = name
	seas.EndDate = end
	_middleDt := seas.BeginDate.Add(seas.EndDate.Sub(seas.BeginDate) / 2)
	seas.SolarTimes, err = weather.GetSunriseAndSunset(_middleDt, lat, lon)
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}

	//seas.Settings = GetDefaultBlockSettings(seas)
	seas.FirstTeeTime = seas.SolarTimes.Sunrise
	seas.LastTeeTime = seas.SolarTimes.Sunset
	seas.Gap = time.Minute * 10
	seas.IsOpen = true
	_weekdayCost := 60
	_weekendCost := 79
	_morningOverrideslots := 8
	_eveningOverrideSlots := 50
	_morningDiscount := 10
	_afternoonDiscount := 10
	_morningEnd := seas.FirstTeeTime.Add(seas.Gap * time.Duration(_morningOverrideslots))
	_eveningStart := seas.FirstTeeTime.Add(seas.Gap * time.Duration(_eveningOverrideSlots))
	//write the code to add price overrides
	seas.DefaultSettings = append(seas.DefaultSettings, DetailedBlockSettings{
		Type: int(WeekdayMorning), Name: "Weekday Morning", BeginOverride: seas.FirstTeeTime,
		EndOverride: _morningEnd,
		Price:       float32(_weekdayCost) - float32(_morningDiscount), IsAvail: true,
	})
	seas.DefaultSettings = append(seas.DefaultSettings, DetailedBlockSettings{
		Type: int(WeekdayMidday), Name: "Weekday Midday", BeginOverride: _morningEnd.Add(time.Minute),
		EndOverride: _eveningStart.Add(time.Minute * -1),
		Price:       float32(_weekdayCost), IsAvail: true,
	})
	seas.DefaultSettings = append(seas.DefaultSettings, DetailedBlockSettings{
		Type: int(WeekdayAfternoon), Name: "Weekday Afternoon", BeginOverride: _eveningStart,
		EndOverride: seas.LastTeeTime.Add(time.Minute),
		Price:       float32(_weekdayCost) - float32(_afternoonDiscount), IsAvail: true,
	})
	seas.DefaultSettings = append(seas.DefaultSettings, DetailedBlockSettings{
		Type: int(WeekendMorning), Name: "Weekend Morning", BeginOverride: seas.FirstTeeTime,
		EndOverride: _morningEnd,
		Price:       float32(_weekendCost) - float32(_morningDiscount), IsAvail: true,
	})
	seas.DefaultSettings = append(seas.DefaultSettings, DetailedBlockSettings{
		Type: int(WeekendMidday), Name: "Weekend Midday", BeginOverride: _morningEnd.Add(time.Minute),
		EndOverride: _eveningStart.Add(time.Minute * -1),
		Price:       float32(_weekendCost), IsAvail: true,
	})
	seas.DefaultSettings = append(seas.DefaultSettings, DetailedBlockSettings{
		Type: int(WeekendAfternoon), Name: "Weekend Afternoon", BeginOverride: _eveningStart,
		EndOverride: seas.LastTeeTime.Add(time.Minute),
		Price:       float32(_weekendCost) - float32(_afternoonDiscount), IsAvail: true,
	})

	return seas
}

func InitNewSeason(year int) []Season {

	var s []Season

	//setup the Four Seasons
	_seasons := helper.GetSeasonsMetereological(year)
	for season, dts := range _seasons {
		if dts[1].After(time.Now()) {
			seas := NewSeason(year, season, dts[0], dts[1])
			//reservationBlock := NewReservationBlock(seas)
			seas.Save()
			s = append(s, seas)
		}
	}

	return s

}
func (s *Season) GetTimeDetails(_date time.Time, _time time.Time) *DetailedBlockSettings {

	for _, setting := range s.DefaultSettings {
		if setting.MatchesType(_date, _time) {
			if setting.BeginOverride.Before(_time) && setting.EndOverride.After(_time) {
				//return the sestting
				return &setting
			} else if setting.BeginOverride.Equal(_time) || setting.EndOverride.Equal(_time) {
				return &setting
			}
		}
	}
	return nil
}
func (s *Season) Save() error {
	_strOut, err := db.Instance.SaveStruct(s, "Season")
	if err != nil {
		fmt.Println(err)
		return err
	}
	s.ID = _strOut

	for i := range s.DefaultSettings {
		_dID, err := s.DefaultSettings[i].Save()
		if err != nil {
			return err
		}
		s.DefaultSettings[i].ID = _dID
		//add Relationship
		err = db.Instance.SaveRelationship(db.Relation{NodeN: "Season", NodeX: "DetailedBlockSettings", NodeNID: s.ID, NodeXID: s.DefaultSettings[i].ID, Name: "HAS_SETTINGS"})
		if err != nil {
			return err
		}
	}
	return nil
}

func GetSeasons(_time time.Time) ([]Season, error) {
	//get the tee times
	query := fmt.Sprintf(`MATCH (n:Season) 
		MATCH (n)-[r:HAS_SETTINGS]->(x:DetailedBlockSettings)
		WITH n, COLLECT(x{.*}) AS defaultSettings
		WHERE date(n.endDate) >= date("%s")
		RETURN n{.*, defaultSettings} as data`, _time.Format(time.DateOnly))

	dayWithRelationships, err := db.Instance.QueryForMap(query, nil) // depth of 2
	if err != nil {
		log.Printf("Error querying with relationships: %v", err)
		return nil, err
	}
	if len(dayWithRelationships) == 0 {
		return nil, nil
	}

	seasonOut := convertMapsToSeasons(dayWithRelationships)
	return seasonOut, nil
}

func GetSeason(_time time.Time) (*Season, error) {
	//get the tee times
	query := fmt.Sprintf(`MATCH (n:Season) 
		MATCH (n)-[r:HAS_SETTINGS]->(x:DetailedBlockSettings)
		WITH n, COLLECT(x{.*}) AS defaultSettings
		WHERE date(n.endDate) >= date("%s") AND date(n.beginDate) <= date("%s")
		RETURN n{.*, defaultSettings} as data`, _time.Format(time.DateOnly), _time.Format(time.DateOnly))

	dayWithRelationships, err := db.Instance.QueryForMap(query, nil) // depth of 2
	if err != nil {
		log.Printf("Error querying with relationships: %v", err)
		return nil, err
	}
	if len(dayWithRelationships) == 0 {
		return nil, nil
	}
	fmt.Println("Application timezone set to:", time.Local.String())
	seasonOut := convertMapsToSeasons(dayWithRelationships)
	if len(seasonOut) > 0 {
		return &seasonOut[0], nil
	}
	return nil, nil
}

// convertMapsToSeasons manually converts []map[string]any to []Season preserving time.Time locations
func convertMapsToSeasons(maps []map[string]any) []Season {
	var seasons []Season

	for _, m := range maps {
		var season Season

		// Convert basic fields
		if id, ok := m["id"].(string); ok {
			season.ID = id
		}
		if year, ok := m["year"].(int64); ok {
			season.Year = int(year)
		}
		if name, ok := m["name"].(string); ok {
			season.Name = name
		}
		if beginDate, ok := m["beginDate"].(time.Time); ok {
			season.BeginDate = beginDate
		}
		if endDate, ok := m["endDate"].(time.Time); ok {
			season.EndDate = endDate
		}
		if firstTeeTime, ok := m["firstTeeTime"].(time.Time); ok {
			season.FirstTeeTime = firstTeeTime
		}
		if lastTeeTime, ok := m["lastTeeTime"].(time.Time); ok {
			season.LastTeeTime = lastTeeTime
		}
		if gap, ok := m["gap"].(int64); ok {
			season.Gap = time.Duration(gap)
		}
		if isOpen, ok := m["isOpen"].(bool); ok {
			season.IsOpen = isOpen
		}

		// Convert defaultSettings array
		if defaultSettings, ok := m["defaultSettings"].([]interface{}); ok {
			for _, setting := range defaultSettings {
				if settingMap, ok := setting.(map[string]interface{}); ok {
					season.DefaultSettings = append(season.DefaultSettings, mapToBlockDetails(settingMap))
				}
			}
		}

		seasons = append(seasons, season)
	}

	return seasons
}
