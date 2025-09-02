package components

import (
	"bigfoot/golf/common/models/auth"
	"fmt"
	"time"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

// HamburgerNav component
type HamburgerNav struct {
	app.Compo
	isOpen   bool
	isAuth   bool
	authResp auth.AuthResponse
}

func (h *HamburgerNav) OnMount(ctx app.Context) {
	ctx.ObserveState(StateKey, &h.authResp).
		OnChange(func() {
			fmt.Println("auth state changed at", time.Now())
			ctx.Dispatch(func(ctx app.Context) {
				if h.authResp.AuthLevel > auth.NoAuthLevel {
					h.isAuth = true
				} else {
					h.isAuth = false
				}
			})
		})
}
func (h *HamburgerNav) getOpenStyle() string {
	if h.isOpen {
		return "active"
	}
	return ""
}
func (h *HamburgerNav) Render() app.UI {

	navItems := []NavItem{
		{Path: "/", Icon: "🏠", Label: "Home"},
		{Path: "/teetimes", Icon: "🏌️", Label: "Tee Times"},
		{Path: "/login", Icon: "👤", Label: "Login", AuthPath: "/account", AuthLabel: "Account"},
		{AuthPath: "/bookings", Icon: "📅", AuthLabel: "My Reservations"},
		{AuthPath: "/agent", Icon: "🤖", AuthLabel: "AI Assistant"},
		{Path: "/about", Icon: "⛳️", Label: "About"},
		{AuthPath: "/admin", Icon: "✏️", AuthLabel: "Admin", IsAdmin: true},
	}
	return app.Div().
		Class("hamburger-nav").
		Body(
			// Header with hamburger button
			app.Header().
				Class("nav-header").
				Body(
					app.Div().
						Class("nav-brand").
						Body(
							app.H1().Text("🏌️ Bigfoot Golf Course"),
						),
					app.Button().
						Class("hamburger-btn").
						Class(h.getOpenStyle()).
						OnClick(h.toggleMenu).
						Body(
							app.Span().Class("hamburger-line"),
							app.Span().Class("hamburger-line"),
							app.Span().Class("hamburger-line"),
						),
				),
			// Overlay
			app.Div().
				Class("nav-overlay").
				Class(h.getOpenStyle()).
				OnClick(h.closeMenu),
			// Slide-out menu
			app.Nav().
				Class("nav-menu").
				Class(h.getOpenStyle()).
				Body(
					app.Div().
						Class("nav-menu-header").
						Body(
							app.H2().Text("Menu"),
							app.Button().
								Class("nav-close-btn").
								OnClick(h.closeMenu).
								Text("×"),
						),
					app.Ul().
						Class("nav-menu-list").
						Body(
							app.Range(navItems).Slice(func(i int) app.UI {
								item := navItems[i]
								if item.AuthPath != "" && h.authResp.AuthLevel > auth.NoAuthLevel {
									if item.IsAdmin && !h.authResp.User.IsAdmin {
										return nil
									}
									return app.Li().Body(
										app.A().
											//Href("#home").
											Text(fmt.Sprintf("%s  %s", item.Icon, item.AuthLabel)).
											OnClick(func(ctx app.Context, e app.Event) {
												h.clickNav(ctx, item.AuthPath)
											}),
									)
								} else if item.Path != "" {
									return app.Li().Body(
										app.A().
											//Href("#home").
											Text(fmt.Sprintf("%s  %s", item.Icon, item.Label)).
											OnClick(func(ctx app.Context, e app.Event) {
												h.clickNav(ctx, item.Path)
											}),
									)
								} else {
									return nil
								}
							},
							),
						),
				),
		)
}

func (h *HamburgerNav) toggleMenu(ctx app.Context, e app.Event) {
	h.isOpen = !h.isOpen
	h.Render()
	//h.Update()
}

func (h *HamburgerNav) closeMenu(ctx app.Context, e app.Event) {
	h.isOpen = false
	h.Render()
	//h.Update()
}
func (h *HamburgerNav) clickNav(ctx app.Context, route string) {
	h.isOpen = false
	//h.Render()
	fmt.Println("Route: ", route)
	ctx.Navigate(route)
	//h.Update()
}
