package courses

import (
	"bigfoot/golf/common/models/account"
	"bigfoot/golf/common/models/db"
	"bigfoot/golf/common/models/teetimes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Course struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	TeeTimeURL string    `json:"teeTimeURL"`
	Headers    []string  `json:"headers"` //Key:Value Collection
	Params     []string  `json:"params"`  //Key:Value Collection
	Active     bool      `json:"active"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func NewCourse(name, teeTimeUrl string, headers, params []string) Course {
	var c Course
	c.Name = name
	c.TeeTimeURL = teeTimeUrl
	c.Headers = headers
	c.Params = params
	c.Active = true
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()
	c.Save()
	return c
}

func (c *Course) Save() error {
	_strOut, err := db.Instance.SaveStruct(c, "Course")
	if err != nil {
		fmt.Println(err)
		return err
	}
	c.ID = _strOut
	return nil
}

// ShItemPrice represents pricing information for tee times
type ShItemPrice struct {
	ItemGuid          string  `json:"itemGuid"`
	ShItemCode        string  `json:"shItemCode"`
	ItemCode          string  `json:"itemCode"`
	Price             float64 `json:"price"`
	TaxInclusivePrice float64 `json:"taxInclusivePrice"`
	TaxCode           string  `json:"taxCode"`
	ItemDesc          string  `json:"itemDesc"`
	ClassCode         string  `json:"classCode"`
	RateCode          string  `json:"rateCode"`
	CurrentPrice      float64 `json:"currentPrice"`
	PriceType         int     `json:"priceType"`
	PriceTypeName     string  `json:"priceTypeName"`
	IsForceOnlineRate bool    `json:"isForceOnlineRate"`
}

// ExternalTeeTime represents the structure from the external API
type ExternalTeeTime struct {
	TeeSheetID                 int           `json:"teeSheetId"`
	StartTime                  string        `json:"startTime"`
	CourseTimeID               int           `json:"courseTimeId"`
	StartingTee                int           `json:"startingTee"`
	CrossOverTeeSheetID        int           `json:"crossOverTeeSheetId"`
	Participants               int           `json:"participants"`
	CourseID                   int           `json:"courseId"`
	CourseDate                 string        `json:"courseDate"`
	DefaultRateCode            string        `json:"defaultRateCode"`
	TeeTypeID                  int           `json:"teeTypeId"`
	Holes                      int           `json:"holes"`
	DefaultHoles               int           `json:"defaultHoles"`
	SiteID                     int           `json:"siteId"`
	CourseName                 string        `json:"courseName"`
	CourseNameIncludeCrossOver string        `json:"courseNameIncludeCrossOver"`
	ShItemPrices               []ShItemPrice `json:"shItemPrices"`
	HolesDisplay               string        `json:"holesDisplay"`
	PlayersDisplay             string        `json:"playersDisplay"`
	MinPlayer                  int           `json:"minPlayer"`
	MaxPlayer                  int           `json:"maxPlayer"`
	AvailableParticipantNo     []int         `json:"availableParticipantNo"`
	IsContain9HoleItems        bool          `json:"isContain9HoleItems"`
	IsContain18HoleItems       bool          `json:"isContain18HoleItems"`
	DefaultClassCode           string        `json:"defaultClassCode"`
	TeeSuffix                  string        `json:"teeSuffix"`
}

// GetRealtime fetches real-time tee time data from the course's API
func (c *Course) GetRealtime(searchDate time.Time) ([]byte, error) {
	// Parse the base URL
	baseURL := c.TeeTimeURL
	if baseURL == "" {
		return nil, fmt.Errorf("course TeeTimeURL is not configured")
	}

	// Build query parameters from c.Params
	queryParams := url.Values{}
	for _, param := range c.Params {
		parts := strings.SplitN(param, ":", 2)
		if len(parts) == 2 {
			key := parts[0]
			value := parts[1]

			// Replace date placeholder if present
			if strings.Contains(strings.ToLower(key), "date") {
				value = searchDate.Format("Mon Jan 2 2006")
			}
			queryParams.Add(key, value)
		}
	}

	// Add searchDate to query params if not already present
	if !queryParams.Has("searchDate") {
		queryParams.Add("searchDate", searchDate.Format("Mon Jan 2 2006"))
	}

	// Build full URL with query parameters
	fullURL := baseURL
	if len(queryParams) > 0 {
		fullURL = baseURL + "?" + queryParams.Encode()
	}

	// Create HTTP request
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers from c.Headers
	for _, header := range c.Headers {
		parts := strings.SplitN(header, ":", 2)
		if len(parts) == 2 {
			req.Header.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}

	// Set default headers if not already set
	if req.Header.Get("accept") == "" {
		req.Header.Set("accept", "application/json, text/plain, */*")
	}
	if req.Header.Get("user-agent") == "" {
		req.Header.Set("user-agent", "Mozilla/5.0 (Linux; Android 13; Pixel 7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Mobile Safari/537.36")
	}

	// Make the HTTP request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tee times: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse the external API response
	var externalResp []ExternalTeeTime
	if err := json.Unmarshal(body, &externalResp); err != nil {
		return nil, fmt.Errorf("failed to parse external API response: %w", err)
	}

	// Transform to our internal format
	reservedDay, err := c.transformToReservedDay(externalResp, searchDate)
	if err != nil {
		return nil, fmt.Errorf("failed to transform tee times: %w", err)
	}

	// Marshal as array of ReservedDay
	result := []teetimes.ReservedDay{reservedDay}
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return resultJSON, nil
}

// transformToReservedDay converts external API tee times to our internal format
func (c *Course) transformToReservedDay(externalTeeTimes []ExternalTeeTime, searchDate time.Time) (teetimes.ReservedDay, error) {
	var reservedDay teetimes.ReservedDay
	reservedDay.Day = searchDate
	reservedDay.Times = []teetimes.Reservation{}

	for idx, extTime := range externalTeeTimes {
		// Parse the startTime (RFC3339 format: "2025-10-01T15:10:00")
		teeTime, err := time.Parse("2006-01-02T15:04:05", extTime.StartTime)
		if err != nil {
			fmt.Printf("Warning: failed to parse time %s: %v\n", extTime.StartTime, err)
			continue
		}

		// Convert to local timezone
		teeTime = time.Date(
			teeTime.Year(),
			teeTime.Month(),
			teeTime.Day(),
			teeTime.Hour(),
			teeTime.Minute(),
			0, 0,
			db.TimeLocation,
		)

		// Extract the price (prefer 18-hole green fee, fallback to first price)
		var price float32 = 0
		if len(extTime.ShItemPrices) > 0 {
			// Look for 18-hole green fee first
			for _, item := range extTime.ShItemPrices {
				if item.ShItemCode == "GreenFee18" || strings.Contains(strings.ToLower(item.ItemDesc), "18") {
					price = float32(item.Price)
					break
				}
			}
			// If no 18-hole found, use first price
			if price == 0 {
				price = float32(extTime.ShItemPrices[0].Price)
			}
		}

		// Create empty player list to represent available spots
		players := []account.User{}

		reservation := teetimes.Reservation{
			TeeTime: teeTime,
			Slot:    int64(idx + 1),
			Price:   price,
			Players: players,
		}
		reservedDay.Times = append(reservedDay.Times, reservation)

	}

	return reservedDay, nil
}

// parseTimeString parses various time formats and combines with the search date
func ParseTimeString(timeStr string, date time.Time) (time.Time, error) {
	timeStr = strings.TrimSpace(timeStr)

	// Try common formats
	formats := []string{
		"3:04 PM",
		"03:04 PM",
		"15:04",
		"3:04PM",
		"03:04PM",
	}

	var parsedTime time.Time
	var err error
	for _, format := range formats {
		parsedTime, err = time.Parse(format, timeStr)
		if err == nil {
			// Success - combine with date
			return time.Date(
				date.Year(),
				date.Month(),
				date.Day(),
				parsedTime.Hour(),
				parsedTime.Minute(),
				0, 0,
				db.TimeLocation,
			), nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time string: %s", timeStr)
}

// GetAll retrieves all courses from the database
func GetAll() ([]Course, error) {
	query := `MATCH (n:Course) RETURN n as data`
	results, err := db.Instance.QueryForMap(query, nil)
	if err != nil {
		return nil, err
	}

	var courses []Course
	for _, result := range results {
		var course Course
		if id, ok := result["id"].(string); ok {
			course.ID = id
		}
		if name, ok := result["name"].(string); ok {
			course.Name = name
		}
		if teeTimeURL, ok := result["teeTimeURL"].(string); ok {
			course.TeeTimeURL = teeTimeURL
		}
		if headers, ok := result["headers"].([]interface{}); ok {
			course.Headers = make([]string, len(headers))
			for i, h := range headers {
				if hs, ok := h.(string); ok {
					course.Headers[i] = hs
				}
			}
		}
		if params, ok := result["params"].([]interface{}); ok {
			course.Params = make([]string, len(params))
			for i, p := range params {
				if ps, ok := p.(string); ok {
					course.Params[i] = ps
				}
			}
		}
		if active, ok := result["active"].(bool); ok {
			course.Active = active
		}
		if createdAt, ok := result["createdAt"].(time.Time); ok {
			course.CreatedAt = createdAt
		}
		if updatedAt, ok := result["updatedAt"].(time.Time); ok {
			course.UpdatedAt = updatedAt
		}
		courses = append(courses, course)
	}

	return courses, nil
}

// GetByID retrieves a single course by ID
func GetByID(courseID string) (*Course, error) {
	query := `MATCH (n:Course {id: $id}) RETURN n as data`
	params := map[string]any{"id": courseID}

	results, err := db.Instance.QueryForMap(query, params)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("course not found with id: %s", courseID)
	}

	result := results[0]
	var course Course
	if id, ok := result["id"].(string); ok {
		course.ID = id
	}
	if name, ok := result["name"].(string); ok {
		course.Name = name
	}
	if teeTimeURL, ok := result["teeTimeURL"].(string); ok {
		course.TeeTimeURL = teeTimeURL
	}
	if headers, ok := result["headers"].([]interface{}); ok {
		course.Headers = make([]string, len(headers))
		for i, h := range headers {
			if hs, ok := h.(string); ok {
				course.Headers[i] = hs
			}
		}
	}
	if params, ok := result["params"].([]interface{}); ok {
		course.Params = make([]string, len(params))
		for i, p := range params {
			if ps, ok := p.(string); ok {
				course.Params[i] = ps
			}
		}
	}
	if active, ok := result["active"].(bool); ok {
		course.Active = active
	}
	if createdAt, ok := result["createdAt"].(time.Time); ok {
		course.CreatedAt = createdAt
	}
	if updatedAt, ok := result["updatedAt"].(time.Time); ok {
		course.UpdatedAt = updatedAt
	}

	return &course, nil
}

// Remove sets the Active field to false instead of deleting the course
func (c *Course) Remove() error {
	c.Active = false
	c.UpdatedAt = time.Now()
	return c.Save()
}
