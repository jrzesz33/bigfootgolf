package admin

import (
	"bigfoot/golf/common/models/courses"
	"encoding/json"
	"fmt"
	"net/http"
)

// GetCourses returns all courses
func GetCourses(w http.ResponseWriter, r *http.Request) {
	allCourses, err := courses.GetAll()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving courses: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allCourses)
}

// SaveCourse creates or updates a course
func SaveCourse(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID         string   `json:"id"`
		Name       string   `json:"name"`
		TeeTimeURL string   `json:"teeTimeURL"`
		Headers    []string `json:"headers"`
		Params     []string `json:"params"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Name == "" || req.TeeTimeURL == "" {
		http.Error(w, "Name and TeeTimeURL are required", http.StatusBadRequest)
		return
	}

	var course courses.Course
	if req.ID != "" {
		// Update existing course
		existingCourse, err := courses.GetByID(req.ID)
		if err != nil {
			http.Error(w, "Course not found", http.StatusNotFound)
			return
		}
		course = *existingCourse
		course.Name = req.Name
		course.TeeTimeURL = req.TeeTimeURL
		course.Headers = req.Headers
		course.Params = req.Params
	} else {
		// Create new course
		course = courses.NewCourse(req.Name, req.TeeTimeURL, req.Headers, req.Params)
	}

	// Save course
	if err := course.Save(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save course: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"course":  course,
	})
}

// ToggleCourseActive toggles the active status of a course
func ToggleCourseActive(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CourseID string `json:"courseId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.CourseID == "" {
		http.Error(w, "CourseID is required", http.StatusBadRequest)
		return
	}

	// Get the course
	course, err := courses.GetByID(req.CourseID)
	if err != nil {
		http.Error(w, "Course not found", http.StatusNotFound)
		return
	}

	// Toggle active status
	if course.Active {
		if err := course.Remove(); err != nil {
			http.Error(w, fmt.Sprintf("Failed to deactivate course: %v", err), http.StatusInternalServerError)
			return
		}
	} else {
		course.Active = true
		if err := course.Save(); err != nil {
			http.Error(w, fmt.Sprintf("Failed to activate course: %v", err), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"active":  !course.Active, // Return the new status
	})
}
