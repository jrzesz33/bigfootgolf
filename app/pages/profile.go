package pages

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Profile struct {
	app.Compo
}

func (p *Profile) Render() app.UI {
	return app.Div().
		Class("page profile-page").
		Body(
			app.Div().
				Class("page-header").
				Body(
					app.H1().Text("Profile"),
				),

			app.Div().
				Class("profile-section").
				Body(
					app.Div().
						Class("profile-avatar").
						Text("👤"),
					app.H2().Text("John Golfer"),
					app.P().Text("Handicap: 18"),
				),

			app.Div().
				Class("profile-options").
				Body(
					app.Button().
						Class("profile-option").
						Text("Edit Profile"),
					app.Button().
						Class("profile-option").
						Text("Preferences"),
					app.Button().
						Class("profile-option").
						Text("Payment Methods"),
					app.Button().
						Class("profile-option").
						Text("Help & Support"),
				),
		)
}
