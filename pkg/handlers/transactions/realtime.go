package transactions

import (
	"bigfoot/golf/common/models/courses"
	"encoding/json"
	"net/http"
	"time"
)

type RealtimeRequest struct {
	CourseID   string `json:"courseId"`
	SearchDate string `json:"searchDate"` // Expected format: "2006-01-02"
}

func GetRealtimeTeeTimes(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	var req RealtimeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate course ID
	if req.CourseID == "" {
		http.Error(w, "courseId is required", http.StatusBadRequest)
		return
	}

	// Parse the search date
	searchDate, err := time.Parse(time.DateOnly, req.SearchDate)
	if err != nil {
		http.Error(w, "Invalid date format. Expected YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	// Get the course from the database
	course, err := courses.GetByID(req.CourseID)
	if err != nil {
		http.Error(w, "Course not found", http.StatusNotFound)
		return
	}

	// Check if course is active
	if !course.Active {
		http.Error(w, "Course is not active", http.StatusBadRequest)
		return
	}

	// Call GetRealtime to fetch tee times
	teeTimesData, err := course.GetRealtime(searchDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(teeTimesData)
}
