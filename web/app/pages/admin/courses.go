package admin

import (
	"bigfoot/golf/common/models/courses"
	"bigfoot/golf/web/app/clients"
	"bigfoot/golf/web/app/components/form"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type CoursesManagement struct {
	app.Compo
	courses        []courses.Course
	showForm       bool
	editingCourse  *courses.Course
	formName       string
	formURL        string
	headersList    form.KeyValueList
	paramsList     form.KeyValueList
	errorMessage   string
	successMessage string
}

func (c *CoursesManagement) LoadCourses(ctx app.Context) {
	ctx.Async(func() {
		_body, erb := clients.SendPostWithAuth("./admin/courses/list", "")
		if erb.Code != 200 {
			ctx.Dispatch(func(ctx app.Context) {
				c.errorMessage = fmt.Sprintf("Failed to load courses: Status %d", erb.Code)
			})
			return
		}

		var loadedCourses []courses.Course
		err := json.Unmarshal(_body, &loadedCourses)
		if err != nil {
			ctx.Dispatch(func(ctx app.Context) {
				c.errorMessage = fmt.Sprintf("Failed to parse courses: %v", err)
			})
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.courses = loadedCourses
			c.errorMessage = ""
		})
	})
}

func (c *CoursesManagement) Render() app.UI {
	return app.Div().Class("courses-management").Body(
		app.H2().Text("Manage Other Courses"),
		app.P().Text("Add and manage external golf courses for real-time tee time lookups."),

		// Error/Success Messages
		app.If(c.errorMessage != "", func() app.UI {
			return app.Div().Class("message error-message").Text(c.errorMessage)
		}),
		app.If(c.successMessage != "", func() app.UI {
			return app.Div().Class("message success-message").Text(c.successMessage)
		}),

		// Add New Course Button
		app.If(!c.showForm, func() app.UI {
			return app.Button().
				Class("btn btn-primary").
				Text("+ Add New Course").
				OnClick(c.onShowForm)
		}),

		// Course Form
		app.If(c.showForm, func() app.UI {
			return c.renderForm()
		}),

		// Courses List
		app.Div().Class("courses-list").Body(
			app.H3().Text("Existing Courses"),
			app.If(len(c.courses) == 0, func() app.UI {
				return app.P().Text("No courses configured yet.")
			}),
			app.Range(c.courses).Slice(func(i int) app.UI {
				course := c.courses[i]
				return app.Div().Class("course-item").Body(
					app.Div().Class("course-info").Body(
						app.H4().Text(course.Name),
						app.P().Text(fmt.Sprintf("URL: %s", course.TeeTimeURL)),
						app.P().Text(fmt.Sprintf("Active: %v", course.Active)),
						app.P().Text(fmt.Sprintf("Headers: %d", len(course.Headers))),
						app.P().Text(fmt.Sprintf("Params: %d", len(course.Params))),
					),
					app.Div().Class("course-actions").Body(
						app.Button().
							Class("btn btn-secondary").
							Text("Edit").
							DataSet("index", fmt.Sprintf("%d", i)).
							OnClick(c.onEditCourse),
						app.Button().
							Class("btn btn-danger").
							Text(func() string {
								if course.Active {
									return "Deactivate"
								}
								return "Activate"
							}()).
							DataSet("courseId", course.ID).
							OnClick(c.onToggleActive),
					),
				)
			}),
		),
	)
}

func (c *CoursesManagement) renderForm() app.UI {
	return app.Div().Class("course-form").Body(
		app.H3().Text(func() string {
			if c.editingCourse != nil {
				return "Edit Course"
			}
			return "Add New Course"
		}()),

		app.Div().Class("form-group").Body(
			app.Label().For("courseName").Text("Course Name:"),
			app.Input().
				Type("text").
				ID("courseName").
				Value(c.formName).
				Placeholder("Enter course name").
				OnInput(c.onNameInput),
		),

		app.Div().Class("form-group").Body(
			app.Label().For("courseURL").Text("Tee Time URL:"),
			app.Input().
				Type("text").
				ID("courseURL").
				Value(c.formURL).
				Placeholder("https://api.example.com/teetimes").
				OnInput(c.onURLInput),
		),

		app.Div().Class("form-group").Body(
			&c.headersList,
		),

		app.Div().Class("form-group").Body(
			&c.paramsList,
		),

		app.Div().Class("form-actions").Body(
			app.Button().
				Class("btn btn-primary").
				Text("Save Course").
				OnClick(c.onSaveCourse),
			app.Button().
				Class("btn btn-secondary").
				Text("Cancel").
				OnClick(c.onCancelForm),
		),
	)
}

func (c *CoursesManagement) onShowForm(ctx app.Context, e app.Event) {
	e.PreventDefault()
	ctx.Dispatch(func(ctx app.Context) {
		c.showForm = true
		c.editingCourse = nil
		c.formName = ""
		c.formURL = ""
		c.headersList.Label = "Headers"
		c.headersList.Pairs = nil
		c.paramsList.Label = "Parameters"
		c.paramsList.Pairs = nil
		c.errorMessage = ""
		c.successMessage = ""
	})
}

func (c *CoursesManagement) onCancelForm(ctx app.Context, e app.Event) {
	e.PreventDefault()
	ctx.Dispatch(func(ctx app.Context) {
		c.showForm = false
		c.editingCourse = nil
		c.errorMessage = ""
	})
}

func (c *CoursesManagement) onNameInput(ctx app.Context, e app.Event) {
	c.formName = ctx.JSSrc().Get("value").String()
}

func (c *CoursesManagement) onURLInput(ctx app.Context, e app.Event) {
	c.formURL = ctx.JSSrc().Get("value").String()
}

func (c *CoursesManagement) onEditCourse(ctx app.Context, e app.Event) {
	e.PreventDefault()
	indexStr := ctx.JSSrc().Get("dataset").Get("index").String()
	var index int
	fmt.Sscanf(indexStr, "%d", &index)

	if index >= 0 && index < len(c.courses) {
		course := c.courses[index]

		// Convert headers and params to KeyValuePairs
		var headerPairs []form.KeyValuePair
		for _, h := range course.Headers {
			parts := strings.SplitN(h, ":", 2)
			if len(parts) == 2 {
				headerPairs = append(headerPairs, form.KeyValuePair{
					Key:   parts[0],
					Value: parts[1],
				})
			}
		}

		var paramPairs []form.KeyValuePair
		for _, p := range course.Params {
			parts := strings.SplitN(p, ":", 2)
			if len(parts) == 2 {
				paramPairs = append(paramPairs, form.KeyValuePair{
					Key:   parts[0],
					Value: parts[1],
				})
			}
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.editingCourse = &course
			c.formName = course.Name
			c.formURL = course.TeeTimeURL
			c.headersList.Label = "Headers"
			c.headersList.Pairs = headerPairs
			c.paramsList.Label = "Parameters"
			c.paramsList.Pairs = paramPairs
			c.showForm = true
			c.errorMessage = ""
			c.successMessage = ""
		})
	}
}

func (c *CoursesManagement) onSaveCourse(ctx app.Context, e app.Event) {
	e.PreventDefault()

	// Validate inputs
	if c.formName == "" || c.formURL == "" {
		ctx.Dispatch(func(ctx app.Context) {
			c.errorMessage = "Course name and URL are required"
		})
		return
	}

	// Convert KeyValuePairs to "Key:Value" strings
	var headers []string
	for _, pair := range c.headersList.Pairs {
		headers = append(headers, fmt.Sprintf("%s:%s", pair.Key, pair.Value))
	}

	var params []string
	for _, pair := range c.paramsList.Pairs {
		params = append(params, fmt.Sprintf("%s:%s", pair.Key, pair.Value))
	}

	// Build course data
	courseData := map[string]interface{}{
		"name":       c.formName,
		"teeTimeURL": c.formURL,
		"headers":    headers,
		"params":     params,
	}

	if c.editingCourse != nil {
		courseData["id"] = c.editingCourse.ID
	}

	// Send to backend
	ctx.Async(func() {
		jsonData, _ := json.Marshal(courseData)
		_, erb := clients.SendPostWithAuth("./admin/courses/save", string(jsonData))

		if erb.Code != 200 {
			ctx.Dispatch(func(ctx app.Context) {
				c.errorMessage = fmt.Sprintf("Failed to save course: Status %d", erb.Code)
			})
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.successMessage = "Course saved successfully!"
			c.showForm = false
			c.editingCourse = nil
			c.errorMessage = ""
			c.LoadCourses(ctx)
		})
	})
}

func (c *CoursesManagement) onToggleActive(ctx app.Context, e app.Event) {
	e.PreventDefault()
	courseID := ctx.JSSrc().Get("dataset").Get("courseId").String()

	ctx.Async(func() {
		data := map[string]string{"courseId": courseID}
		jsonData, _ := json.Marshal(data)
		_, erb := clients.SendPostWithAuth("./admin/courses/toggle", string(jsonData))

		if erb.Code != 200 {
			ctx.Dispatch(func(ctx app.Context) {
				c.errorMessage = fmt.Sprintf("Failed to toggle course status: Status %d", erb.Code)
			})
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.successMessage = "Course status updated!"
			c.LoadCourses(ctx)
		})
	})
}
