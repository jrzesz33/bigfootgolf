package pages

import (
	"birdsfoot/app/app/components"
	"fmt"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Home struct {
	app.Compo
	components.BaseCo
}

func (h *Home) OnMount(ctx app.Context) {
	fmt.Println("Mount Triggered for Home Page")
	h.GetFromState(ctx)
}

func (h *Home) OnNav(ctx app.Context) {
	fmt.Println("Nav Triggered for Home Page")
}
func (h *Home) OnDismount() {
	fmt.Println("Dismount Triggered for Home Page")
}

func (h *Home) Render() app.UI {
	fmt.Println("Rendering Home Page")

	return app.Div().Body(
		app.Div().Text(fmt.Sprintf("User- %s %s", h.User.FirstName, h.User.LastName)),
		app.Div().
			Class("quick-actions").
			Body(
				app.Button().
					Class("action-btn primary").
					Text("Quick Book").
					OnClick(h.onQuickBook),

				app.Button().
					Class("action-btn secondary").
					Text("Agent Cat").
					OnClick(h.onCourseInfo),
				app.Button().
					Class("action-btn secondary").
					Text("Register").
					OnClick(h.onRegister),
			))
}

func (h *Home) onQuickBook(ctx app.Context, e app.Event) {
	ctx.Navigate("/search")
}

func (h *Home) onCourseInfo(ctx app.Context, e app.Event) {
	// Show course information
	ctx.Navigate("/chat")
}

func (h *Home) onRegister(ctx app.Context, e app.Event) {
	// Show course information
	ctx.Navigate("/register")
}
