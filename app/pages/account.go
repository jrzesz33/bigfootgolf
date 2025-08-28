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
	ctx.GetState(components.StateKey, &h.authResp)
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

func (h *MyAccount) onLogout(ctx app.Context, e app.Event) {

	_state := state.GetAppState(nil)
	_state.Logout()
}
func (h *MyAccount) onProfileEdit(ctx app.Context, e app.Event) {
	ctx.Dispatch(func(ctx app.Context) {
		h.disabledMode = false
	})
}

func (h *MyAccount) onPasswordReset(ctx app.Context, e app.Event) {
	if !h.user.IsVerified {
		ctx.Navigate("/verify")
	} else {
		ctx.Navigate("/changepw")
	}

}
func (h *MyAccount) onSaveProfileClick(_user account.User) {
	_usr, err := updateUser(_user)
	if _usr != nil && err.BError == nil {
		appState := state.GetAppState(nil)
		appState.UpdateUser(_user)
		h.user = _user
		//h.statusMsg = "Profile Updated"
	} else {
		fmt.Println("Error Saving Profile", err)
		h.statusMsg = err.FriendlyMsg()
	}

	h.disabledMode = true
	h.Render()

}

func (h *MyAccount) onCancelClick() {
	h.disabledMode = true
}

func updateUser(_user account.User) (*account.User, models.BError) {
	body, _ := json.Marshal(_user)

	usr, err := clients.SendPostWithAuth("./api/userupdate", string(body))
	if err.BError == nil {
		//success update the state
		var _usr account.User
		err.BError = json.Unmarshal(usr, &_usr)
		return &_usr, err
	}
	return nil, err
}
