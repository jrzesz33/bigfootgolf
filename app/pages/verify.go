package pages

import (
	"birdsfoot/app/app/clients"
	"birdsfoot/app/app/state"
	"birdsfoot/app/models/account"
	"encoding/json"
	"strings"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type VerifyUI struct {
	app.Compo
	User               account.User
	displayValidateBtn bool
	errorMsg           string
	usersCode          string
}

func (h *VerifyUI) OnMount(ctx app.Context) {
	h.displayValidateBtn = false
	as := state.GetAppState(nil)
	//as.Subscribe()

	h.User = as.TokenManager().GetAuth().User
}

func (h *VerifyUI) Render() app.UI {
	return app.Div().
		Body(
			app.Div().Class("errMsg").
				Text(h.errorMsg),
			app.Form().
				OnSubmit(h.verifyClick).
				Body(
					app.Div().Text("Your account requires verification. A code was emailed to the account on file, please enter that code and click Verify."),
					app.Label().For("verifyCode").Text("Verification Code"),
					app.Input().Type("verifyCode").
						ID("verifyCode").
						OnChange(h.ValueTo(&h.usersCode)).
						Required(true),
					app.Div().
						Class("quick-actions").
						Body(
							app.Button().Text("Get Code").Class("action-btn secondary").OnClick(h.getCode).Hidden(h.displayValidateBtn),
							app.Button().Text("Verify Code").Class("action-btn secondary").Type("submit").Hidden(!h.displayValidateBtn),
							//app.Button().Text("Cancel").Class("action-btn secondary").OnClick(h.cancelClick),
						),
				))
}

func (h *VerifyUI) getCode(ctx app.Context, opts app.Event) {
	if h.User.ID == "" {
		h.errorMsg = "Unknown User, please re-login"
		return
	}
	body, _ := json.Marshal(h.User)
	_, err := clients.SendPostWithAuth("./api/verifyreq", string(body))
	if err.BError != nil || err.Code != 200 {
		h.errorMsg = "There was an issue sending the code, please try again later."
		return
	}
	ctx.Dispatch(func(ctx app.Context) {
		h.displayValidateBtn = true
		h.errorMsg = ""
	})

}

func (h *VerifyUI) verifyClick(ctx app.Context, opts app.Event) {
	opts.PreventDefault()
	_codeElm := app.Window().GetElementByID("verifyCode").Get("value").String()
	if len(strings.TrimSpace(_codeElm)) != 6 {
		h.errorMsg = "Please enter a six digit code"
		return
	}
	body, _ := json.Marshal(h.User)
	_, err := clients.SendPostWithAuth("./api/verifyemailcode", string(body))
	if err.BError != nil || err.Code > 200 {
		h.errorMsg = "Incorrect Code, please try again."
		return
	}
	h.User.IsVerified = true
	as := state.GetAppState(nil)
	as.UpdateUser(h.User)
	app.Window().GetElementByID("verifyCode").Set("value", "")
	ctx.Navigate("/")

}
