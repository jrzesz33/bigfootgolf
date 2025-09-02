package weather

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// SolarResults contains all the solar time information
type SolarResults struct {
	Sunrise                   string `json:"sunrise"`
	Sunset                    string `json:"sunset"`
	SolarNoon                 string `json:"solar_noon"`
	DayLength                 string `json:"day_length"`
	CivilTwilightBegin        string `json:"civil_twilight_begin"`
	CivilTwilightEnd          string `json:"civil_twilight_end"`
	NauticalTwilightBegin     string `json:"nautical_twilight_begin"`
	NauticalTwilightEnd       string `json:"nautical_twilight_end"`
	AstronomicalTwilightBegin string `json:"astronomical_twilight_begin"`
	AstronomicalTwilightEnd   string `json:"astronomical_twilight_end"`
}

// SunriseSunsetResponse represents the complete API response
type SunriseSunsetResponse struct {
	Results SolarResults `json:"results"`
	Status  string       `json:"status"`
	TZID    string       `json:"tzid"`
}

// ParsedSolarResults contains parsed time.Time values for easier manipulation
type ParsedSolarResults struct {
	Sunrise                   time.Time     `json:"sunrise"`
	Sunset                    time.Time     `json:"sunset"`
	SolarNoon                 time.Time     `json:"solarNoon"`
	DayLength                 time.Duration `json:"dayLength"`
	CivilTwilightBegin        time.Time     `json:"civilTwilightBegin"`
	CivilTwilightEnd          time.Time     `json:"civilTwilightEnd"`
	NauticalTwilightBegin     time.Time     `json:"nauticalTwilightBegin"`
	NauticalTwilightEnd       time.Time     `json:"nauticalTwilightEnd"`
	AstronomicalTwilightBegin time.Time     `json:"astronomicalTwilightBegin"`
	AstronomicalTwilightEnd   time.Time     `json:"astronomicalTwilightEnd"`
}

// ParseTimes converts string times to time.Time objects using the provided timezone
func (r *SolarResults) ParseTimes(location *time.Location) (*ParsedSolarResults, error) {
	parseTime := func(timeStr string) (time.Time, error) {
		// Parse time in format "3:04:05 PM"
		return time.ParseInLocation("3:04:05 PM", timeStr, location)
	}

	parseDuration := func(durationStr string) (time.Duration, error) {
		// Parse duration in format "15:09:06"
		// Construct the duration string by replacing colons with appropriate unit suffixes
		durString := durationStr[0:2] + "h" + durationStr[3:5] + "m" + durationStr[6:8] + "s"

		return time.ParseDuration(durString)
	}

	sunrise, err := parseTime(r.Sunrise)
	if err != nil {
		return nil, fmt.Errorf("parsing sunrise: %w", err)
	}

	sunset, err := parseTime(r.Sunset)
	if err != nil {
		return nil, fmt.Errorf("parsing sunset: %w", err)
	}

	solarNoon, err := parseTime(r.SolarNoon)
	if err != nil {
		return nil, fmt.Errorf("parsing solar noon: %w", err)
	}

	dayLength, err := parseDuration(r.DayLength)
	if err != nil {
		return nil, fmt.Errorf("parsing day length: %w", err)
	}

	civilTwilightBegin, err := parseTime(r.CivilTwilightBegin)
	if err != nil {
		return nil, fmt.Errorf("parsing civil twilight begin: %w", err)
	}

	civilTwilightEnd, err := parseTime(r.CivilTwilightEnd)
	if err != nil {
		return nil, fmt.Errorf("parsing civil twilight end: %w", err)
	}

	nauticalTwilightBegin, err := parseTime(r.NauticalTwilightBegin)
	if err != nil {
		return nil, fmt.Errorf("parsing nautical twilight begin: %w", err)
	}

	nauticalTwilightEnd, err := parseTime(r.NauticalTwilightEnd)
	if err != nil {
		return nil, fmt.Errorf("parsing nautical twilight end: %w", err)
	}

	astronomicalTwilightBegin, err := parseTime(r.AstronomicalTwilightBegin)
	if err != nil {
		return nil, fmt.Errorf("parsing astronomical twilight begin: %w", err)
	}

	astronomicalTwilightEnd, err := parseTime(r.AstronomicalTwilightEnd)
	if err != nil {
		return nil, fmt.Errorf("parsing astronomical twilight end: %w", err)
	}

	return &ParsedSolarResults{
		Sunrise:                   sunrise,
		Sunset:                    sunset,
		SolarNoon:                 solarNoon,
		DayLength:                 dayLength,
		CivilTwilightBegin:        civilTwilightBegin,
		CivilTwilightEnd:          civilTwilightEnd,
		NauticalTwilightBegin:     nauticalTwilightBegin,
		NauticalTwilightEnd:       nauticalTwilightEnd,
		AstronomicalTwilightBegin: astronomicalTwilightBegin,
		AstronomicalTwilightEnd:   astronomicalTwilightEnd,
	}, nil
}

func GetSunriseAndSunset(dt time.Time, lat float32, lon float32) (*ParsedSolarResults, error) {

	//https://api.sunrise-sunset.org/json?lat=40.745152&lng=-79.665367&tzid=America/New_York&date=2025-06-20
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	url := fmt.Sprintf("https://api.sunrise-sunset.org/json?lat=%v&lng=%v&tzid=%s&date=%s", lat, lon, dt.Location().String(), dt.Format("2006-01-02"))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set User-Agent header as required by weather.gov API
	//req.Header.Set("User-Agent", "WeatherApp/1.0 (your-email@example.com)")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weather data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	var response SunriseSunsetResponse
	err = json.Unmarshal([]byte(body), &response)
	if err != nil {
		fmt.Printf("Error unmarshaling JSON: %v\n", err)
		return nil, err
	}

	fmt.Printf("Status: %s\n", response.Status)
	fmt.Printf("Timezone: %s\n", response.TZID)
	fmt.Printf("Sunrise: %s\n", response.Results.Sunrise)
	fmt.Printf("Sunset: %s\n", response.Results.Sunset)

	// Parse times with timezone
	location, err := time.LoadLocation(response.TZID)
	if err != nil {
		fmt.Printf("Error loading location: %v\n", err)
		return nil, err
	}

	parsed, err := response.Results.ParseTimes(location)
	if err != nil {
		fmt.Printf("Error parsing times: %v\n", err)
		return nil, err
	}
	return parsed, nil
}
