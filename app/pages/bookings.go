package pages

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Bookings struct {
	app.Compo
}

func (b *Bookings) Render() app.UI {
	return app.Div().
		Class("page bookings-page").
		Body(
			app.Div().
				Class("page-header").
				Body(
					app.H1().Text("My Bookings"),
				),

			app.Div().
				Class("booking-list").
				Body(
					app.Div().
						Class("booking-card").
						Body(
							app.H3().Text("Upcoming Tee Time"),
							app.P().Text("Saturday, June 15th at 10:30 AM"),
							app.P().Text("4 Players - $220 total"),
							app.Div().
								Class("booking-actions").
								Body(
									app.Button().
										Class("btn secondary").
										Text("Modify"),
									app.Button().
										Class("btn danger").
										Text("Cancel"),
								),
						),
				),
		)
}
