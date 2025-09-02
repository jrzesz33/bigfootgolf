package components

import (
	"bigfoot/golf/common/models/auth"
	"bigfoot/golf/web/app/state"
	"fmt"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

// StateKey is the key used to store application state
const StateKey string = "birdstate"

type Layout struct {
	app.Compo
	Page      app.UI
	Header    string
	PageLevel auth.AuthLevel
	eventChan chan state.StateEvent
	userLevel auth.AuthLevel
}

func (aw *Layout) OnMount(ctx app.Context) {

	var authResp auth.AuthResponse
	ctx.GetState(StateKey, &authResp)
	if authResp.Token == "" {
		fmt.Println("No Local Storage")
	}

	appState := state.GetAppState(&authResp)
	aw.eventChan = appState.Subscribe()
	aw.userLevel = appState.TokenManager().GetAuth().AuthLevel

	if aw.Header == "Home" && authResp.User.FirstName != "" {
		aw.Header = "Welcome Back " + authResp.User.FirstName
	}

	if aw.PageLevel > aw.userLevel {
		if aw.userLevel == auth.NoAuthLevel {
			//nav to Login, user not logged in
			ctx.Navigate("/login")
		} else {
			//does not have access to page.... handle step up or other messaging
			fmt.Println("DOES NOT HAVE AUTH LEVEL ACCESS TO PAGE")
			ctx.Navigate("/")
		}
	}
	// Start listening for auth events
	go aw.listenForAuthEvents(ctx)
}

func (aw *Layout) OnDismount() {
	if aw.eventChan != nil {
		state.GetAppState(nil).Unsubscribe(aw.eventChan)
	}
}

func (aw *Layout) listenForAuthEvents(ctx app.Context) {
	for event := range aw.eventChan {
		fmt.Println("EVENT FIRED: ", event.Type)
		switch event.Type {
		case "auth_failed":
			aw.ClearAuthState(ctx)
			// Update UI on main thread
			fmt.Println("Auth Failed")

		case "login_success":
			aw.UpdateAuthState(ctx, event)
			ctx.Navigate("/")

		case "logout":
			aw.ClearAuthState(ctx)
			ctx.Navigate("/login")

		case "update_user":
			aw.UpdateAuthState(ctx, event)
		case "token_refreshed":
			aw.RefrestAuthState(ctx, event)
		default:
			fmt.Println("Event NOT FOUND: ", event.Type)
		}

	}
}
func (aw *Layout) UpdateAuthState(ctx app.Context, event state.StateEvent) {
	if _auth, ok := event.Data.(auth.AuthResponse); ok {
		fmt.Println("Updating User to ", _auth.User.LastName)
		aw.userLevel = _auth.AuthLevel
		ctx.SetState(StateKey, _auth).Persist().Broadcast()
	} else {
		fmt.Println("Error with Saving State")
	}
}

func (aw *Layout) RefrestAuthState(ctx app.Context, event state.StateEvent) {
	if _auth, ok := event.Data.(auth.AuthResponse); ok {
		aw.userLevel = _auth.AuthLevel
		var authResp auth.AuthResponse
		ctx.GetState(StateKey, &authResp)
		_auth.User = authResp.User
		ctx.SetState(StateKey, _auth).Persist().Broadcast()
	} else {
		fmt.Println("Error with Saving State")
	}
}

func (aw *Layout) ClearAuthState(ctx app.Context) {
	aw.userLevel = auth.NoAuthLevel
	var _auth auth.AuthResponse
	ctx.SetState(StateKey, _auth).Persist().Broadcast()
}

func (aw *Layout) Render() app.UI {

	return app.Div().
		Class("app-container").
		Body(
			app.Main().
				Class("main-content").
				Body(
					&HamburgerNav{},
					app.Div().
						Class("page").
						Body(
							app.Div().
								Class("page-header").
								Body(
									app.H2().Text(aw.Header),
								),
							aw.Page,
							&Footer{},
						),
				),
		)
}
