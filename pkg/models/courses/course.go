package courses

import (
	"bigfoot/golf/common/models/db"
	"fmt"
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

// GetRealtime fetches real-time tee time data from the course's API
func (c *Course) GetRealtime(searchDate time.Time) ([]byte, error) {
	// TODO: Implement HTTP client to call c.TeeTimeURL with c.Headers and c.Params
	// This will make an external API call to fetch real-time tee times
	// For now, returning a placeholder error
	return nil, fmt.Errorf("GetRealtime not yet implemented - requires HTTP client to call %s", c.TeeTimeURL)
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
