package weather

import (
	"math/rand"
	"time"
)

// SimpleWeatherData represents simplified weather information for MCP
type SimpleWeatherData struct {
	Date         string  `json:"date"`
	Temperature  float64 `json:"temperature"`
	Condition    string  `json:"condition"`
	WindSpeed    float64 `json:"wind_speed"`
	Humidity     int     `json:"humidity"`
	ChanceOfRain int     `json:"chance_of_rain"`
}

// GetWeatherForecast returns weather forecast for a given date
func GetWeatherForecast(date time.Time) (*SimpleWeatherData, error) {
	// In a real implementation, this would call an external weather API
	// For now, returning mock data
	
	// Generate semi-random but consistent weather based on date
	r := rand.New(rand.NewSource(date.Unix()))
	
	conditions := []string{"Sunny", "Partly Cloudy", "Cloudy", "Light Rain", "Clear"}
	
	weather := &SimpleWeatherData{
		Date:         date.Format("2006-01-02"),
		Temperature:  60 + r.Float64()*30, // 60-90°F
		Condition:    conditions[r.Intn(len(conditions))],
		WindSpeed:    r.Float64() * 20,  // 0-20 mph
		Humidity:     40 + r.Intn(40),   // 40-80%
		ChanceOfRain: r.Intn(30),        // 0-30%
	}
	
	// If it's "Light Rain", increase chance of rain
	if weather.Condition == "Light Rain" {
		weather.ChanceOfRain = 60 + r.Intn(30)
	}
	
	return weather, nil
}

// GetCurrentConditions returns current weather conditions
func GetCurrentConditions() (*SimpleWeatherData, error) {
	return GetWeatherForecast(time.Now())
}