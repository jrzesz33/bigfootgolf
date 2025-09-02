package pages

import (
	"bigfoot/golf/app/clients"
	"bigfoot/golf/app/components"
	"bigfoot/golf/common/models/auth"
	"encoding/json"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type PwResetPage struct {
	app.Compo
	userID     string
	newPW      string
	validatePW string
	errorMsg   string
}

func (p *PwResetPage) OnMount(ctx app.Context) {
	var ath auth.AuthResponse
	ctx.GetState(components.StateKey, &ath)
	if ath.User.ID == "" {
		ctx.Navigate("/login")
	}
	//as.Subscribe()
	p.userID = ath.User.ID
	if !ath.User.IsVerified {
		ctx.Navigate("/verify")
	}
}

func (s *PwResetPage) Render() app.UI {
	_obj := app.Div().
		Body(
			app.Form().Body(
				app.Div().Text(s.errorMsg).Class("errMsg"),
				app.Label().For("password").Text("Password"),
				app.Input().Type("password").ID("password").
					AutoComplete(false).
					Required(true).
					Attr("inputmode", "password"),
				app.Label().For("password2").Text("Repeat Password"),
				app.Input().Type("password").ID("password2").
					AutoComplete(false).
					Required(true).
					Attr("inputmode", "password"),
				app.Div().
					Class("quick-actions").
					Body(
						app.Button().Text("Reset Password").Class("action-btn secondary").OnClick(s.resetClick),
					)))

	return _obj
}
func (p *PwResetPage) resetClick(ctx app.Context, opts app.Event) {
	opts.PreventDefault()
	p.newPW = app.Window().GetElementByID("password").Get("value").String()
	p.validatePW = app.Window().GetElementByID("password2").Get("value").String()
	if p.validatePW != p.newPW {
		p.errorMsg = "The passwords do not match"
		return
	}
	_mapData := make(map[string]string)
	_mapData["id"] = p.userID
	_mapData["password"] = p.newPW

	body, _ := json.Marshal(_mapData)

	_, err := clients.SendPostWithAuth("./api/resetapw", string(body))
	if err.BError == nil {
		//success update the state
		ctx.Navigate("/")
		//return &_usr, err
	} else {
		p.errorMsg = err.BError.Error()
	}

}
