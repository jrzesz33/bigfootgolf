// app/routes/routes.go
package routes

import (
	"bigfoot/golf/common/models/auth"
	"bigfoot/golf/web/app/components"
	"bigfoot/golf/web/app/pages"
	"bigfoot/golf/web/app/pages/admin"
	"log"
	"time"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

func RegisterRoutes() {
	// intialize the local time
	localTime := time.Now()
	log.Println("User's local time:", localTime.Format(time.RFC3339))

	// Public routes
	app.Route("/", publicRoute(&pages.Home{}))
	app.Route("/search", func() app.Composer { return &components.Layout{Page: &pages.Search{}} })
	app.Route("/bookings", func() app.Composer { return &components.Layout{Page: &pages.Bookings{}, PageLevel: auth.LoginLevel} })
	app.Route("/chat", func() app.Composer { return &components.Layout{Page: &pages.ChatAgent{}} })
	app.Route("/login", func() app.Composer { return &components.Layout{Page: &pages.Login{}} })
	app.Route("/register", func() app.Composer { return &components.Layout{Page: &pages.Register{}} })
	app.Route("/teetimes", func() app.Composer { return &components.Layout{Page: &pages.AvailTimes{}} })

	// Authenticated routes
	app.Route("/account", func() app.Composer { return &components.Layout{Page: &pages.MyAccount{}, PageLevel: auth.LoginLevel} })
	app.Route("/agent", func() app.Composer { return &components.Layout{Page: &pages.Agent{}, PageLevel: auth.LoginLevel} })
	app.Route("/verify", func() app.Composer { return &components.Layout{Page: &pages.VerifyUI{}, PageLevel: auth.LoginLevel} })
	app.Route("/changepw", func() app.Composer { return &components.Layout{Page: &pages.PwResetPage{}, PageLevel: auth.LoginLevel} })

	// Admin routes
	app.Route("/admin", func() app.Composer { return &components.Layout{Page: &admin.Administer{}, PageLevel: auth.LoginLevel} })
	//RegisterProtectedRoutes()

}

func publicRoute(_page app.UI) func() app.Composer {
	return func() app.Composer { return &components.Layout{Page: _page} }
}
