package form

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type DropDown struct {
	app.Compo
	menuClass  string
	menuOpen   bool
	MenuMap    []map[string]string
	MenuSelect func(string)
}

func (h *DropDown) OnMount(ctx app.Context) {

}

func (h *DropDown) Render() app.UI {
	return app.Div().Class("dropdown-container").Body(
		app.Div().Class("dropdown-trigger", h.menuClass).Role("button").TabIndex(0).Aria("haspopup", true).Aria("expanded", h.menuOpen).Body(
			app.Span().Text("Select an Option"),
			app.Div().Class("dropdown-arrow"),
		).OnClick(h.onMenuClick),
		app.Div().Class("dropdown-menu", h.menuClass).Body(
			app.Ul().Role("menu").Body(
				app.Range(h.MenuMap).Slice(func(i int) app.UI {
					return app.Li().Body(
						app.A().
							Href("#").
							Role("menuitem").
							//Attr("data-seasonID", h.seasons[i].Name).
							DataSet("val", h.MenuMap[i]["value"]).
							Text(h.MenuMap[i]["name"]).
							OnClick(h.onMenuSelect))
				}),
			),
		))
}

func (h *DropDown) onMenuClick(ctx app.Context, e app.Event) {
	e.PreventDefault()
	e.StopImmediatePropagation()
	ctx.Dispatch(func(ctx app.Context) {
		if h.menuOpen {
			h.menuOpen = false
			h.menuClass = ""
		} else {
			h.menuOpen = true
			h.menuClass = "active"
		}
	})

}
func (h *DropDown) onMenuSelect(ctx app.Context, opts app.Event) {
	opts.PreventDefault()
	//disable the form
	ctx.Dispatch(func(ctx app.Context) {
		h.menuOpen = false
		h.menuClass = ""
		//load the season info
		v := ctx.JSSrc().Get("dataset").Get("val").String()
		if h.MenuSelect != nil {
			h.MenuSelect(v)
		}

	})
}
