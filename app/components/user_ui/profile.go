package userui

import (
	"birdsfoot/app/models/account"
	"fmt"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type ProfileUI struct {
	app.Compo
	User            account.User
	SaveProfile     func(account.User)
	RegisterUser    func(account.User)
	Cancelled       func()
	IsDisabled      bool
	IsNewAccount    bool
	showRegisterBtn bool
	ErrorMsg        string
	buttonTxt       string
}

func (h *ProfileUI) OnMount(ctx app.Context) {
	if h.IsNewAccount {
		h.IsDisabled = false
		h.showRegisterBtn = true
		h.buttonTxt = "Register Account"

	} else {
		h.IsDisabled = true
		h.showRegisterBtn = false
		h.buttonTxt = "Save"
	}

}

func (h *ProfileUI) Render() app.UI {
	return app.Div().
		Class().
		Body(
			app.Div().Class("errMsg").
				Text(h.ErrorMsg),
			app.Form().
				OnSubmit(h.profileClick).
				Body(
					app.Label().For("email").Text("Email"),
					app.Input().Type("email").ID("email").AutoComplete(true).
						Value(h.User.Email).
						Attr("inputmode", "email").
						Required(true).
						Disabled(!h.IsNewAccount),
					app.Label().For("first").Text("First Name"),
					app.Input().Type("text").ID("first").
						Value(h.User.FirstName).
						Required(true).
						Disabled(h.IsDisabled),
					app.Label().For("last").Text("Last Name"),
					app.Input().Type("text").ID("last").
						Disabled(h.IsDisabled).
						Required(true).
						Value(h.User.LastName),
					app.Label().For("phone").Text("Phone"),
					app.Input().Type("tel").ID("phone").
						Disabled(h.IsDisabled).
						Attr("autocomplete", "tel").
						Value(h.User.Phone).
						Required(true).
						Attr("inputmode", "tel"),
					app.Label().For("password").Text("Password").Hidden(!h.IsNewAccount),
					app.Input().Type("password").ID("password").
						Disabled(!h.IsNewAccount).
						Hidden(!h.IsNewAccount).
						AutoComplete(false).
						Required(true).
						Attr("inputmode", "password"),
					app.Div().
						Class("quick-actions").
						Body(
							app.Button().Text(h.buttonTxt).Class("action-btn secondary").Type("submit").Hidden(h.IsDisabled),
							app.If(!h.IsNewAccount, func() app.UI {
								return app.Button().ID("saveBtn").Text("Cancel").Class("action-btn secondary").OnClick(h.cancelClick).Hidden(h.IsDisabled)
							}),
						),
				))
}

func (h *ProfileUI) profileClick(ctx app.Context, opts app.Event) {
	opts.PreventDefault()
	//disable the form
	h.IsDisabled = true
	ctx.Update()

	fmt.Println(h.User.Email)

	h.User.Email = app.Window().GetElementByID("email").Get("value").String()
	h.User.FirstName = app.Window().GetElementByID("first").Get("value").String()
	h.User.LastName = app.Window().GetElementByID("last").Get("value").String()
	h.User.Phone = app.Window().GetElementByID("phone").Get("value").String()
	h.User.Password = app.Window().GetElementByID("password").Get("value").String()

	//fmt.Println("clicked profile button", h.User)

	if h.User.Email == "" || h.User.FirstName == "" || h.User.LastName == "" {
		strLog := fmt.Sprintf("User: %s, Name:%s, Pass,%s, Phone:%s", h.User.Email, h.User.FirstName, h.User.Password, h.User.Phone)
		h.ErrorMsg = "Please ensure required fields are filled in."
		fmt.Println(strLog)
		return
	}
	fmt.Println("Registering User to Server")
	if h.IsNewAccount {
		if h.RegisterUser != nil {
			h.RegisterUser(h.User)
		}
		return
	}
	if h.SaveProfile != nil {
		h.SaveProfile(h.User)
	}
}

func (h *ProfileUI) cancelClick(ctx app.Context, opts app.Event) {
	//add validation logic here
	if h.Cancelled != nil {
		h.Cancelled()
	}
}
