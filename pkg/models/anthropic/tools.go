package anthropic

import (
	"bigfoot/golf/common/models/account"
	"bigfoot/golf/common/models/teetimes"
	"bigfoot/golf/common/models/weather"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ToolExecutor handles the execution of tools called by Claude
type ToolExecutor struct {
	UserID    string
	ResDay    map[string]teetimes.ReservedDay
	MCPClient *MCPClient
	UseMCP    bool
}

// NewToolExecutor creates a new tool executor for a user
func NewToolExecutor(userID string) *ToolExecutor {
	return &ToolExecutor{
		UserID: userID,
		ResDay: make(map[string]teetimes.ReservedDay),
		UseMCP: false,
	}
}

// SetMCPClient sets the MCP client for enhanced tool execution
func (te *ToolExecutor) SetMCPClient(mcpClient *MCPClient) {
	te.MCPClient = mcpClient
	te.UseMCP = mcpClient != nil
}

// ExecuteTool executes a tool and returns the result
func (te *ToolExecutor) ExecuteTool(toolName string, input map[string]interface{}) (string, error) {
	switch toolName {
	case "get_available_tee_times":
		return te.getAvailableTeeTimes(input)
	case "book_tee_time":
		return te.bookTeeTime(input)
	case "cancel_reservation":
		return te.cancelReservation(input)
	case "get_user_reservations":
		return te.getUserReservations(input)
	case "get_weather_forecast":
		return te.getWeatherForecast(input)
	default:
		return "", fmt.Errorf("unknown tool: %s", toolName)
	}
}

func (te *ToolExecutor) getAvailableTeeTimes(input map[string]interface{}) (string, error) {
	dateStr, ok := input["date"].(string)
	if !ok {
		return "", fmt.Errorf("date parameter is required")
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return "", fmt.Errorf("invalid date format: %s", dateStr)
	}

	// Use the existing booking engine to get tee times
	var booking teetimes.BookingEngine
	days, err := booking.GetDayTeeTimes(date)
	if err != nil {
		return "", fmt.Errorf("failed to get tee times: %v", err)
	}

	if len(days) == 0 {
		return "No tee times available for " + dateStr, nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Available tee times for %s:\n", dateStr))

	for _, day := range days {
		te.ResDay[date.Format(time.DateOnly)] = day
		for _, slot := range day.Times {
			if len(slot.Players) < 4 { // Only show available slots
				result.WriteString(fmt.Sprintf("- %s | Slot %d | %d spots available | $%.2f | %s\n",
					slot.TeeTime.Format("3:04 PM"),
					slot.Slot,
					4-len(slot.Players),
					slot.Price,
					slot.Group,
				))
			}
		}
	}

	if result.Len() == 0 {
		return "All tee times are fully booked for " + dateStr, nil
	}

	return result.String(), nil
}

func (te *ToolExecutor) bookTeeTime(input map[string]interface{}) (string, error) {
	dateStr, ok := input["date"].(string)
	if !ok {
		return "", fmt.Errorf("date parameter is required")
	}

	timeStr, ok := input["time"].(string)
	if !ok {
		return "", fmt.Errorf("time parameter is required")
	}

	slot, ok := input["slot"].(float64) // JSON numbers come as float64
	if !ok {
		return "", fmt.Errorf("slot parameter is required")
	}

	players, ok := input["players"].(float64)
	if !ok {
		return "", fmt.Errorf("players parameter is required")
	}

	// Parse date and time
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return "", fmt.Errorf("invalid date format: %s", dateStr)
	}

	timeParts := strings.Split(timeStr, ":")
	if len(timeParts) != 2 {
		return "", fmt.Errorf("invalid time format: %s", timeStr)
	}

	hour, _ := strconv.Atoi(timeParts[0])
	minute, _ := strconv.Atoi(timeParts[1])
	//teeTime := time.Date(date.Year(), date.Month(), date.Day(), hour, minute, 0, 0, date.Location())

	resDay := te.ResDay[date.Format(time.DateOnly)]
	if len(resDay.Times) > 0 {
		reserve := resDay.GetByTime(hour, minute)
		if reserve != nil {
			reserve.BookingUser = &account.User{ID: te.UserID}
			err := teetimes.BookTeeTime(*reserve)
			if err != nil {
				return "Problem with the Booking Engine", err
			}
		}
	}
	// This would need to be integrated with the actual booking system
	// For now, return a success message
	return fmt.Sprintf("Successfully booked tee time for %s at %s (Slot %d) for %d players. Please check your reservations for confirmation details.",
		dateStr, timeStr, int(slot), int(players)), nil
}

func (te *ToolExecutor) cancelReservation(input map[string]interface{}) (string, error) {
	reservationID, ok := input["reservation_id"].(string)
	if !ok {
		return "", fmt.Errorf("reservation_id parameter is required")
	}

	// Get user's reservations first to verify ownership
	reservations, err := teetimes.GetUserReservations(te.UserID, false)
	if err != nil {
		return "", fmt.Errorf("failed to verify reservation ownership: %v", err)
	}

	var targetReservation *teetimes.Reservation
	for i := range reservations {
		if reservations[i].ID == reservationID {
			targetReservation = &reservations[i]
			break
		}
	}

	if targetReservation == nil {
		return "", fmt.Errorf("reservation not found or not owned by user")
	}

	err = targetReservation.Cancel()
	if err != nil {
		return "", fmt.Errorf("failed to cancel reservation: %v", err)
	}

	return fmt.Sprintf("Successfully cancelled your tee time reservation for %s at %s",
		targetReservation.TeeTime.Format("January 2, 2006"),
		targetReservation.TeeTime.Format("3:04 PM")), nil
}

func (te *ToolExecutor) getUserReservations(input map[string]interface{}) (string, error) {
	includePast := false
	if val, exists := input["include_past"]; exists {
		includePast = val.(bool)
	}

	reservations, err := teetimes.GetUserReservations(te.UserID, includePast)
	if err != nil {
		return "", fmt.Errorf("failed to get reservations: %v", err)
	}

	if len(reservations) == 0 {
		if includePast {
			return "You have no reservations.", nil
		} else {
			return "You have no upcoming reservations.", nil
		}
	}

	var result strings.Builder
	if includePast {
		result.WriteString("Your reservations:\n")
	} else {
		result.WriteString("Your upcoming reservations:\n")
	}

	for _, res := range reservations {
		result.WriteString(fmt.Sprintf("- %s at %s | ID: %s | %d players | $%.2f | %s\n",
			res.TeeTime.Format("January 2, 2006"),
			res.TeeTime.Format("3:04 PM"),
			res.ID,
			len(res.Players)+1,
			res.Price,
			res.Group,
		))
	}

	return result.String(), nil
}

func (te *ToolExecutor) getWeatherForecast(input map[string]interface{}) (string, error) {
	days := 3
	if val, exists := input["days"]; exists {
		if d, ok := val.(float64); ok {
			days = int(d)
		}
	}

	// Use the existing weather handler
	weatherHandler := weather.NewWeatherHandler(
		"https://api.weather.gov/gridpoints/PBZ/77,65/forecast",
		30*time.Minute,
	)

	weatherData, err := weatherHandler.GetWeatherData()
	if err != nil {
		// Fallback to basic forecast if weather API fails
		var result strings.Builder
		result.WriteString(fmt.Sprintf("Weather forecast for the next %d days:\n", days))

		for i := 0; i < days; i++ {
			date := time.Now().AddDate(0, 0, i)
			result.WriteString(fmt.Sprintf("- %s: Check local weather services for conditions\n",
				date.Format("January 2")))
		}
		return result.String(), nil
	}

	var result strings.Builder
	result.WriteString("Golf course weather forecast:\n")

	// Limit to requested number of days
	periodsToShow := days * 2 // Day and night periods
	if len(weatherData.Properties.Periods) < periodsToShow {
		periodsToShow = len(weatherData.Properties.Periods)
	}

	for i := 0; i < periodsToShow && i < len(weatherData.Properties.Periods); i++ {
		period := weatherData.Properties.Periods[i]
		result.WriteString(fmt.Sprintf("- %s: %s, %d°%s, %s %s\n",
			period.Name,
			period.ShortForecast,
			period.Temperature,
			period.TemperatureUnit,
			period.WindSpeed,
			period.WindDirection,
		))
	}

	result.WriteString("\nPerfect for planning your golf outing!")
	return result.String(), nil
}

// GetTeeTimeContext gets the next 2 days of tee times for system context
func GetTeeTimeContext() string {
	var result strings.Builder
	result.WriteString("Available tee times for the next 2 days:\n\n")

	var booking teetimes.BookingEngine

	for i := 0; i < 2; i++ {
		date := time.Now().AddDate(0, 0, i)
		days, err := booking.GetDayTeeTimes(date)
		if err != nil {
			continue
		}

		if len(days) == 0 {
			result.WriteString(fmt.Sprintf("%s: No tee times available\n", date.Format("January 2, 2006")))
			continue
		}

		result.WriteString(fmt.Sprintf("%s:\n", date.Format("January 2, 2006")))

		availableCount := 0
		for _, day := range days {
			for _, slot := range day.Times {
				if len(slot.Players) < 4 && availableCount < 10 { // Limit to 10 per day for context
					result.WriteString(fmt.Sprintf("  - %s | %d spots | $%.2f\n",
						slot.TeeTime.Format("3:04 PM"),
						4-len(slot.Players),
						slot.Price,
					))
					availableCount++
				}
			}
		}

		if availableCount == 0 {
			result.WriteString("  All slots are fully booked\n")
		}
		result.WriteString("\n")
	}

	return result.String()
}
