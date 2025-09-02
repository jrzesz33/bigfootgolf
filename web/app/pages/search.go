package pages

import (
	"fmt"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Search struct {
	app.Compo
	selectedDate string
	//playerCount  int
	timeSlots []TimeSlot
}

type TimeSlot struct {
	Time  string
	Price string
	ID    string
}

func (s *Search) Render() app.UI {
	_obj := app.Div().
		Class("time-slots").
		Body(
			app.H3().Text("Available Times"),
			app.Range(s.timeSlots).Slice(func(i int) app.UI {
				slot := s.timeSlots[i]
				return app.Div().
					Class("time-slot").
					Body(
						app.Div().
							Class("slot-time").
							Text(slot.Time),
						app.Div().
							Class("slot-price").
							Text(slot.Price),
						app.Button().
							Class("btn secondary").
							Text("Book").
							OnClick(s.onBookSlot),
					)
			}),
		)
	fmt.Println(_obj)
	return app.Div().
		Class("page search-page").
		Body(
			app.Div().
				Class("page-header").
				Body(
					app.H1().Text("Find Tee Times"),
				),

			app.Div().
				Class("search-form").
				Body(
					app.Div().
						Class("form-group").
						Body(
							app.Label().Text("Date"),
							app.Input().
								Type("date").
								Class("form-input").
								Value(s.selectedDate).
								OnChange(s.onDateChange),
						),

					app.Div().
						Class("form-group").
						Body(
							app.Label().Text("Players"),
							app.Select().
								Class("form-select").
								Body(
									app.Option().Value("1").Text("1 Player"),
									app.Option().Value("2").Text("2 Players"),
									app.Option().Value("3").Text("3 Players"),
									app.Option().Value("4").Text("4 Players"),
								).
								OnChange(s.onPlayerChange),
						),

					app.Button().
						Class("btn primary full-width").
						Text("Search Available Times").
						OnClick(s.onSearch),
				),

			app.If(len(s.timeSlots) > 0, nil),
		)
}

func (s *Search) onDateChange(ctx app.Context, e app.Event) {
	s.selectedDate = ctx.JSSrc().Get("value").String()
}

func (s *Search) onPlayerChange(ctx app.Context, e app.Event) {
	// Handle player count change
}

func (s *Search) onSearch(ctx app.Context, e app.Event) {
	// Mock data - replace with actual API call
	s.timeSlots = []TimeSlot{
		{Time: "8:30 AM", Price: "$45", ID: "1"},
		{Time: "10:15 AM", Price: "$55", ID: "2"},
		{Time: "2:30 PM", Price: "$40", ID: "3"},
	}
	s.Render()
}

func (s *Search) onBookSlot(ctx app.Context, e app.Event) {
	ctx.Navigate("/bookings")
}
