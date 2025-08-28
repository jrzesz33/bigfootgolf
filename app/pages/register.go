package pages

import (
	userui "birdsfoot/app/app/components/user_ui"
	"birdsfoot/app/app/state"
	"birdsfoot/app/models/account"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Register struct {
	app.Compo
	newUser  account.User
	errorMsg string
	//playerCount  int
}

func (s *Register) Render() app.UI {
	_obj := app.Div().
		Body(
			&userui.ProfileUI{
				IsNewAccount: true,
				IsDisabled:   false,
				User:         s.newUser,
				RegisterUser: s.onRegisterClick,
				ErrorMsg:     s.errorMsg,
			},
		)
	return _obj
}

func (r *Register) onRegisterClick(_user account.User) {
	//register the user
	appState := state.GetAppState(nil)
	err := appState.RegisterUser(_user)
	if err.BError != nil {
		r.errorMsg = err.FriendlyMsg()
	}

}
