package pages

import (
	"bigfoot/golf/common/models/auth"
	"bigfoot/golf/web/app/state"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Login struct {
	app.Compo
	newLogin auth.LoginRequest
	errorMsg string
	//playerCount  int
}

func (s *Login) Render() app.UI {
	_obj := app.Div().
		Body(
			app.Div().Class("errMsg").Text(s.errorMsg),
			app.Label().For("email").Text("Email"),
			app.Input().Type("email").ID("email").AutoComplete(true).
				Attr("inputmode", "email").
				Required(true).
				OnChange(s.ValueTo(&s.newLogin.Email)),

			app.Label().For("password").Text("Password"),
			app.Input().Type("password").ID("password").
				AutoComplete(true).
				OnChange(s.ValueTo(&s.newLogin.Password)).
				Required(true).
				Attr("inputmode", "password"),
			app.Div().
				Class("quick-actions").
				Body(
					app.Button().Text("Login").Class("action-btn primary").OnClick(s.onLoginClick),
					app.Button().Text("Register").Class("action-btn secondary").
						OnClick(func(ctx app.Context, e app.Event) {
							ctx.Navigate("/register")
						}),
				),
		)
	return _obj
}

func (s *Login) onLoginClick(ctx app.Context, e app.Event) {

	appState := state.GetAppState(nil)
	err := appState.Login(s.newLogin)
	if err != nil {
		ctx.Dispatch(func(ctx app.Context) {
			s.errorMsg = err.Error()

		})
	}
}
