package pages

import (
	"birdsfoot/app/app/clients"
	"birdsfoot/app/app/components"
	userui "birdsfoot/app/app/components/user_ui"
	"birdsfoot/app/app/state"
	"birdsfoot/app/models"
	"birdsfoot/app/models/account"
	"birdsfoot/app/models/auth"
	"encoding/json"
	"fmt"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type MyAccount struct {
	app.Compo
	user         account.User
	disabledMode bool
	statusMsg    string
	authResp     auth.AuthResponse
}

func (h *MyAccount) OnMount(ctx app.Context) {
	h.disabledMode = true
	ctx.GetState(components.STATE_KEY, &h.authResp)
	if h.authResp.Token == "" {
		fmt.Println("No User Found")
		ctx.Navigate("/login")
	}
	h.user = h.authResp.User

	//load profile component initially

}
func (h *MyAccount) Render() app.UI {

	return app.Div().
		Body(
			&userui.ProfileUI{
				User:         h.user,
				IsNewAccount: false,
				IsDisabled:   h.disabledMode,
				SaveProfile:  h.onSaveProfileClick,
				Cancelled:    h.onCancelClick,
				ErrorMsg:     h.statusMsg,
			},
			app.Div().
				Class("quick-actions").
				Body(
					app.Button().
						Class("action-btn secondary").
						Text("Edit Profile").
						Hidden(!h.disabledMode).
						OnClick(h.onProfileEdit),
					app.Button().
						Class("action-btn secondary").
						Text("Change Password").
						Hidden(!h.disabledMode).
						OnClick(h.onPasswordReset),
					app.Button().
						Class("action-btn secondary").
						Text("Logout").
						Hidden(!h.disabledMode).
						OnClick(h.onLogout),
				),
		)
}

func (a *MyAccount) onLogout(ctx app.Context, e app.Event) {

	_state := state.GetAppState(nil)
	_state.Logout()
}
func (a *MyAccount) onProfileEdit(ctx app.Context, e app.Event) {
	ctx.Dispatch(func(ctx app.Context) {
		a.disabledMode = false
	})
}

func (a *MyAccount) onPasswordReset(ctx app.Context, e app.Event) {
	if !a.user.IsVerified {
		ctx.Navigate("/verify")
	} else {
		ctx.Navigate("/changepw")
	}

}
func (a *MyAccount) onSaveProfileClick(_user account.User) {
	_usr, err := updateUser(_user)
	if _usr != nil && err.BError == nil {
		appState := state.GetAppState(nil)
		appState.UpdateUser(_user)
		a.user = _user
		//a.statusMsg = "Profile Updated"
	} else {
		fmt.Println("Error Saving Profile", err)
		a.statusMsg = err.FriendlyMsg()
	}

	a.disabledMode = true
	a.Render()

}

func (a *MyAccount) onCancelClick() {
	a.disabledMode = true
}

func updateUser(_user account.User) (*account.User, models.BError) {
	body, _ := json.Marshal(_user)

	usr, err := clients.SendPostWithAuth("./api/userupdate", string(body))
	if err.BError == nil {
		//success update the state
		var _usr account.User
		err.BError = json.Unmarshal(usr, &_usr)
		return &_usr, err
	} else {
		return nil, err
	}
}
