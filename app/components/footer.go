package components

import (
	"birdsfoot/app/app/clients"
	"birdsfoot/app/models/auth"
	"birdsfoot/app/models/weather"
	"encoding/json"
	"fmt"
	"time"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

const LOGGING_LEVEL string = "debug"

type Footer struct {
	app.Compo
	authResp       auth.AuthResponse
	isAuth         bool
	currentWeather weather.WeatherData
	temp           int
	wind           string
	forecast       string
}

type NavItem struct {
	Path      string
	AuthPath  string
	Icon      string
	Label     string
	AuthLabel string
	IsAdmin   bool
}

func (n *Footer) OnMount(ctx app.Context) {
	if LOGGING_LEVEL == "debug" {
		ctx.ObserveState(STATE_KEY, &n.authResp).
			OnChange(func() {
				fmt.Println("auth state changed in footer at", time.Now())
				ctx.Dispatch(func(ctx app.Context) {
					if n.authResp.AuthLevel > auth.NoAuthLevel {
						n.isAuth = true
					} else {
						n.isAuth = false
					}
				})
			})
	}
	ctx.GetState("birdwthr", &n.currentWeather)
	if len(n.currentWeather.Properties.Periods) == 0 {
		fmt.Println("No Local Weather")
		_resp, err := clients.SendGetReq("./papi/weather")
		if err != nil {
			fmt.Println(err)
			return
		}
		err = json.Unmarshal(_resp, &n.currentWeather)
		if err != nil {
			fmt.Println("error unmarshalling weather")
		}
		ctx.SetState("birdwthr", n.currentWeather).ExpiresIn(time.Minute * 10)
	}
	if len(n.currentWeather.Properties.Periods) > 0 {
		_current := n.currentWeather.Properties.Periods[0]
		n.temp = _current.Temperature
		n.wind = fmt.Sprintf("%s winds %s", _current.WindSpeed, _current.WindDirection)
		n.forecast = _current.DetailedForecast
	}
}
func (n *Footer) Render() app.UI {

	return app.Div().
		Body(
			app.Div().
				Class("weather-widget").
				Body(
					app.H3().Text("Today's Weather at Birdsfoot"),
					app.P().Text("A Perfect Day for Golf!"),
					app.P().Style("font-weight", "bold").Text("SCATTER GOLF"),
					app.Div().
						Class("weather-info").
						Body(
							app.Span().Text(fmt.Sprintf("☀️ %d°F", n.temp)),
							app.Span().Text(fmt.Sprintf("💨 %s", n.wind)),
						),
					app.P().
						Style("font-size", "13").
						Style("margin-top", "10").
						Style("font-style", "italic").
						Text(n.forecast),
				),
			app.If(LOGGING_LEVEL == "debug", func() app.UI {
				return app.Div().
					Style("font-family", "monospace").
					Style("font-size", "10px").
					Style("margin-top", "15px").
					Body(
						app.P().Text(fmt.Sprintf("Auth Level: %d", int(n.authResp.AuthLevel))),
						app.P().Text(fmt.Sprintf("Token Expiration: %v", n.authResp.ExpiresIn)),
						app.P().Text(fmt.Sprintf("Refresh Token: %s", n.authResp.RefreshToken)),
						app.P().Text(fmt.Sprintf("Access Token: %s", n.authResp.Token)),
						app.P().Text(fmt.Sprintf("User : %s %s, Email: %s", n.authResp.User.FirstName, n.authResp.User.LastName, n.authResp.User.Email)),
						app.P().Text(fmt.Sprintf("Phone: %s, Verified: %v", n.authResp.User.Phone, n.authResp.User.IsVerified)),
					)
			}),
		)
}
