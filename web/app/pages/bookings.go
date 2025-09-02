package pages

import (
	"bigfoot/golf/app/components"
	"bigfoot/golf/common/models/auth"
	"bigfoot/golf/common/models/teetimes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Bookings struct {
	app.Compo
	reservations []teetimes.Reservation
	showPast     bool
	loading      bool
	error        string
	authResp     auth.AuthResponse
}

func (b *Bookings) OnMount(ctx app.Context) {
	// Observe authentication state
	ctx.ObserveState(components.StateKey, &b.authResp).OnChange(func() {
		ctx.Dispatch(func(ctx app.Context) {
			if b.authResp.AuthLevel > auth.NoAuthLevel {
				b.loadReservations(ctx)
			}
		})
	})

	// Load initial reservations if already authenticated
	if b.authResp.AuthLevel > auth.NoAuthLevel {
		b.loadReservations(ctx)
	}
}

func (b *Bookings) loadReservations(ctx app.Context) {
	b.loading = true
	b.error = ""

	url := "/api/reservations"
	if b.showPast {
		url += "?includePast=true"
	}

	go func() {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			ctx.Dispatch(func(ctx app.Context) {
				b.error = "Failed to create request"
				b.loading = false
			})
			return
		}

		req.Header.Set("Authorization", "Bearer "+b.authResp.Token)
		req.Header.Set("X-User-ID", b.authResp.User.ID)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			ctx.Dispatch(func(ctx app.Context) {
				b.error = "Failed to load reservations"
				b.loading = false
			})
			return
		}
		defer resp.Body.Close()

		var reservations []teetimes.Reservation
		if err := json.NewDecoder(resp.Body).Decode(&reservations); err != nil {
			ctx.Dispatch(func(ctx app.Context) {
				b.error = "Failed to parse reservations"
				b.loading = false
			})
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			b.reservations = reservations
			b.loading = false
		})
	}()
}

func (b *Bookings) cancelReservation(ctx app.Context, reservationID string) {
	ctx.Dispatch(func(ctx app.Context) {
		b.loading = true
	})

	go func() {
		payload := map[string]string{"reservationId": reservationID}
		jsonPayload, _ := json.Marshal(payload)

		req, err := http.NewRequest("POST", "/api/reservations/cancel",
			strings.NewReader(string(jsonPayload)))
		if err != nil {
			ctx.Dispatch(func(ctx app.Context) {
				b.error = "Failed to cancel reservation"
				b.loading = false
			})
			return
		}

		req.Header.Set("Authorization", "Bearer "+b.authResp.Token)
		req.Header.Set("X-User-ID", b.authResp.User.ID)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			ctx.Dispatch(func(ctx app.Context) {
				b.error = "Failed to cancel reservation"
				b.loading = false
			})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			ctx.Dispatch(func(ctx app.Context) {
				b.loadReservations(ctx)
			})
		} else {
			ctx.Dispatch(func(ctx app.Context) {
				b.error = "Failed to cancel reservation"
				b.loading = false
			})
		}
	}()
}

func (b *Bookings) togglePastReservations(ctx app.Context, e app.Event) {
	b.showPast = !b.showPast
	b.loadReservations(ctx)
}

func (b *Bookings) Render() app.UI {
	if b.authResp.AuthLevel <= auth.NoAuthLevel {
		return app.Div().
			Class("page bookings-page").
			Body(
				app.Div().
					Class("page-header").
					Body(
						app.H1().Text("My Reservations"),
						app.P().Text("Please log in to view your reservations."),
					),
			)
	}

	return app.Div().
		Class("page bookings-page").
		Body(
			app.Div().
				Class("page-header").
				Body(
					app.H1().Text("My Reservations"),
					app.Div().
						Class("reservation-controls").
						Body(
							app.Button().
								Class("btn secondary").
								Text(func() string {
									if b.showPast {
										return "Show Future Only"
									}
									return "Show Past Year"
								}()).
								OnClick(b.togglePastReservations),
						),
				),

			app.If(b.loading, func() app.UI {
				return app.Div().
					Class("loading").
					Text("Loading reservations...")
			}),

			app.If(b.error != "", func() app.UI {
				return app.Div().
					Class("error").
					Text(b.error)
			}),

			app.If(!b.loading && b.error == "", func() app.UI {
				return b.renderReservations()
			}),
		)
}

func (b *Bookings) renderReservations() app.UI {
	if len(b.reservations) == 0 {
		return app.Div().
			Class("no-reservations").
			Body(
				app.P().Text("No reservations found."),
				app.A().
					Href("/teetimes").
					Class("btn primary").
					Text("Book a Tee Time"),
			)
	}

	var futureReservations, pastReservations []teetimes.Reservation
	now := time.Now()

	for _, res := range b.reservations {
		if res.TeeTime.After(now) {
			futureReservations = append(futureReservations, res)
		} else {
			pastReservations = append(pastReservations, res)
		}
	}

	return app.Div().
		Class("reservation-list").
		Body(
			// Future reservations
			app.If(len(futureReservations) > 0, func() app.UI {
				return app.Div().
					Class("reservation-section").
					Body(
						app.H2().Text("Upcoming Reservations"),
						app.Range(futureReservations).Slice(func(i int) app.UI {
							return b.renderReservationCard(futureReservations[i], true)
						}),
					)
			}),

			// Past reservations
			app.If(b.showPast && len(pastReservations) > 0, func() app.UI {
				return app.Div().
					Class("reservation-section").
					Body(
						app.H2().Text("Past Reservations"),
						app.Range(pastReservations).Slice(func(i int) app.UI {
							return b.renderReservationCard(pastReservations[i], false)
						}),
					)
			}),
		)
}

func (b *Bookings) renderReservationCard(reservation teetimes.Reservation, canCancel bool) app.UI {
	return app.Div().
		Class("booking-card").
		Body(
			app.Div().
				Class("booking-header").
				Body(
					app.H3().Text(reservation.TeeTime.Format("Monday, January 2, 2006")),
					app.P().
						Class("tee-time").
						Text(reservation.TeeTime.Format("3:04 PM")),
				),

			app.Div().
				Class("booking-details").
				Body(
					app.P().Text(fmt.Sprintf("Players: %d", len(reservation.Players)+1)),
					app.P().Text(fmt.Sprintf("Price: $%.2f", reservation.Price)),
					app.P().Text(fmt.Sprintf("Group: %s", reservation.Group)),
					app.If(len(reservation.Players) > 0, func() app.UI {
						return app.Div().
							Class("players-list").
							Body(
								app.P().Text("Players:"),
								app.Ul().Body(
									app.Range(reservation.Players).Slice(func(i int) app.UI {
										player := reservation.Players[i]
										return app.Li().Text(player.LastName)
									}),
								),
							)
					}),
				),

			app.If(canCancel, func() app.UI {
				return app.Div().
					Class("booking-actions").
					Body(
						app.Button().
							Class("btn danger").
							Text("Cancel").
							OnClick(func(ctx app.Context, e app.Event) {
								if app.Window().Call("confirm", "Are you sure you want to cancel this reservation?").Bool() {
									b.cancelReservation(ctx, reservation.ID)
								}
							}),
					)
			}),
		)
}
