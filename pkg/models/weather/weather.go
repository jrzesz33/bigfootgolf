package weather

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// WeatherData represents the structure of the weather API response
type WeatherData struct {
	Properties struct {
		Periods []struct {
			Number           int    `json:"number"`
			Name             string `json:"name"`
			StartTime        string `json:"startTime"`
			EndTime          string `json:"endTime"`
			IsDaytime        bool   `json:"isDaytime"`
			Temperature      int    `json:"temperature"`
			TemperatureUnit  string `json:"temperatureUnit"`
			WindSpeed        string `json:"windSpeed"`
			WindDirection    string `json:"windDirection"`
			Icon             string `json:"icon"`
			ShortForecast    string `json:"shortForecast"`
			DetailedForecast string `json:"detailedForecast"`
		} `json:"periods"`
	} `json:"properties"`
}

// CachedWeather holds cached weather data with timestamp
type CachedWeather struct {
	Data      WeatherData
	Timestamp time.Time
}

// WeatherHandler handles weather requests with caching
type WeatherHandler struct {
	cache     *CachedWeather
	cacheMux  sync.RWMutex
	apiURL    string
	cacheTime time.Duration
}

// NewWeatherHandler creates a new weather handler with caching
func NewWeatherHandler(apiURL string, cacheTime time.Duration) *WeatherHandler {
	return &WeatherHandler{
		apiURL:    apiURL,
		cacheTime: cacheTime,
	}
}

// isExpired checks if cached data is expired
func (wh *WeatherHandler) isExpired() bool {
	if wh.cache == nil {
		return true
	}
	return time.Since(wh.cache.Timestamp) > wh.cacheTime
}

// fetchWeatherData fetches fresh weather data from the API
func (wh *WeatherHandler) fetchWeatherData() (*WeatherData, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", wh.apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set User-Agent header as required by weather.gov API
	req.Header.Set("User-Agent", "WeatherApp/1.0 (your-email@example.com)")

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

	var weatherData WeatherData
	if err := json.Unmarshal(body, &weatherData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return &weatherData, nil
}

// GetWeatherData returns weather data from cache or fetches fresh data
func (wh *WeatherHandler) GetWeatherData() (*WeatherData, error) {
	wh.cacheMux.RLock()
	if !wh.isExpired() {
		data := wh.cache.Data
		wh.cacheMux.RUnlock()
		return &data, nil
	}
	wh.cacheMux.RUnlock()

	// Cache is expired or doesn't exist, fetch fresh data
	wh.cacheMux.Lock()
	defer wh.cacheMux.Unlock()

	// Double-check in case another goroutine updated the cache
	if !wh.isExpired() {
		return &wh.cache.Data, nil
	}

	weatherData, err := wh.fetchWeatherData()
	if err != nil {
		return nil, err
	}

	// Update cache
	wh.cache = &CachedWeather{
		Data:      *weatherData,
		Timestamp: time.Now(),
	}

	return weatherData, nil
}

// ServeHTTP implements the http.Handler interface
func (wh *WeatherHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	weatherData, err := wh.GetWeatherData()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get weather data: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Add cache info header
	wh.cacheMux.RLock()
	cacheAge := time.Since(wh.cache.Timestamp)
	wh.cacheMux.RUnlock()

	w.Header().Set("X-Cache-Age", fmt.Sprintf("%.0f", cacheAge.Seconds()))
	w.Header().Set("X-Cache-TTL", fmt.Sprintf("%.0f", wh.cacheTime.Seconds()))

	if err := json.NewEncoder(w).Encode(weatherData); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
