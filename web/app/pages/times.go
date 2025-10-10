package pages

import (
	"bigfoot/golf/common/models/account"
	"bigfoot/golf/common/models/auth"
	"bigfoot/golf/common/models/teetimes"
	"bigfoot/golf/web/app/clients"
	"bigfoot/golf/web/app/state"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"
	_ "time/tzdata"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Course struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	TeeTimeURL string   `json:"teeTimeURL"`
	Headers    []string `json:"headers"`
	Params     []string `json:"params"`
	Active     bool     `json:"active"`
}

type AvailTimes struct {
	app.Compo
	selectedDate   time.Time
	selectedCourse string
	courses        []Course
	errorMsg       string
	timeSlots      []teetimes.ReservedDay
	reservSelected *teetimes.Reservation
	players        int
	showPopup      bool
	authResp       *auth.AuthResponse
}

func (p *AvailTimes) OnMount(ctx app.Context) {
	p.selectedDate = time.Now().Local()
	fmt.Println("Selected Date: ", p.selectedDate)
	p.players = 4
	p.selectedCourse = "" // Default to BigFoot (empty string means default)

	// Fetch available courses
	p.fetchCourses()

	// Load tee times
	p.getTeeTimes()
}

func (p *AvailTimes) fetchCourses() {
	resp, err := clients.SendGetReq("papi/courses")
	if err == nil {
		err := json.Unmarshal(resp, &p.courses)
		if err != nil {
			fmt.Println("Error unmarshaling courses:", err)
		}
	} else {
		fmt.Println("Error fetching courses:", err)
	}
}

func (p *AvailTimes) getTeeTimes() {
	// Clear existing time slots
	p.timeSlots = nil

	if p.selectedCourse == "" {
		// Call the default BigFoot tee times endpoint
		_mapData := make(map[string]interface{})
		_mapData["start"] = p.selectedDate

		body, _ := json.Marshal(_mapData)
		resp, err := clients.SendPostWithPayload("./papi/teetimes", string(body))
		if err == nil {
			//success populate the tee times
			erc := json.Unmarshal(resp, &p.timeSlots)
			if erc != nil {
				fmt.Println(erc)
				return
			}
		} else {
			fmt.Println(err)
		}
	} else {
		// Call the realtime tee times endpoint for the selected course
		_mapData := make(map[string]interface{})
		_mapData["courseId"] = p.selectedCourse
		_mapData["searchDate"] = p.selectedDate.Format("2006-01-02")

		body, _ := json.Marshal(_mapData)
		resp, err := clients.SendPostWithPayload("./papi/realtime", string(body))
		if err == nil {
			//success populate the tee times
			erc := json.Unmarshal(resp, &p.timeSlots)
			if erc != nil {
				fmt.Println("Error unmarshaling realtime tee times:", erc)
				return
			}
		} else {
			fmt.Println("Error fetching realtime tee times:", err)
		}
	}
}

func (s *AvailTimes) Render() app.UI {
	if s.showPopup && s.reservSelected != nil {
		return s.renderPopup()
	}
	if len(s.timeSlots) == 0 {
		return app.Div().Body(
			s.renderCourseSelector(),
			app.Div().Text("No Tee Times for " + s.selectedDate.Format("01/02/2006")),
		)
	}

	x := 0

	sort.Slice(s.timeSlots[x].Times, func(i, j int) bool {
		return s.timeSlots[x].Times[i].TeeTime.Before(s.timeSlots[x].Times[j].TeeTime)
	})
	fmt.Println("days and slots ", len(s.timeSlots), " - ", len(s.timeSlots[0].Times))
	_obj := app.Div().Body(
		s.renderCourseSelector(),
		app.Div().Class("fixedTeeHeader").Body(
			app.If(s.selectedDate.After(time.Now().Truncate(24*time.Hour).Local().Add(time.Hour*24)), func() app.UI {
				return app.Div().Class("fixedTeeBtn").Text("⬅️").OnClick(s.onDateBack)
			}),
			app.Div().Class("fixedTeeDate").Text(s.selectedDate.Format("Jan 2 Mon")),
			app.Div().Class("fixedTeeBtn").Text("➡️").OnClick(s.onDateChange),
		),
		app.Div().
			Class("time-slots").
			Body(
				app.Range(s.timeSlots[x].Times).Slice(func(i int) app.UI {
					slot := s.timeSlots[x].Times[i]
					_p := len(slot.Players)

					if _p < 4 {
						_open := strconv.Itoa(4-_p) + "👤"
						return app.Div().
							Class("time-slot").
							Body(
								app.Div().
									Class("slot-time").
									Text(slot.TeeTime.Format("3:04 PM")),
								app.Div().
									Class("slot-spots").
									Text(_open),
								app.Div().
									Class("slot-price").
									Text(fmt.Sprintf("$%.2f", slot.Price)),
								app.Button().
									Class("btn secondary").
									Text("Book").
									Value(slot.Slot).
									OnClick(func(ctx app.Context, e app.Event) {
										s.onSelectReservation(slot, ctx)
									}),
							)
					}
					return nil
				}),
			))

	return _obj
}
func (p *AvailTimes) onSelectReservation(time teetimes.Reservation, ctx app.Context) {
	//get logged in User
	appState := state.GetAppState(nil)
	p.authResp = appState.TokenManager().GetAuth()
	if p.authResp == nil && p.authResp.AuthLevel < auth.LoginLevel {
		ctx.Navigate("./login")
	}
	ctx.Dispatch(func(ctx app.Context) {
		p.reservSelected = &time
		p.showPopup = true
	})
}
func (p *AvailTimes) onBookSlot(ctx app.Context, opts app.Event) {

	time := p.reservSelected

	//add the user to the reservation
	if p.authResp != nil && p.authResp.AuthLevel >= auth.LoginLevel {
		time.BookingUser = &p.authResp.User
		time.Players = append(time.Players, *time.BookingUser)
		for i := 1; i < p.players; i++ {
			time.Players = append(time.Players, account.User{LastName: fmt.Sprintf("Guest %d", i)})
		}
		//book the time
		_slot, _ := json.Marshal(time)
		resp, erb := clients.SendPostWithAuth("./api/bookTime", string(_slot))
		if erb.Code >= 400 || erb.BError != nil {
			fmt.Println("Error Booking Time: ", erb.Code, erb.BError)
			//CACHE THE TIME FOR WHEN ITS WORKING
			ctx.SetState("newRes", _slot)
			//appState.CacheEvent("newRes", time)
			//TODO BUILD MOBILE POPUP TO HIGHLIGHT ERROR
			p.errorMsg = "There was an error booking the time. " + erb.FriendlyMsg()
			p.showPopup = false
			p.reservSelected = nil
			return
		} else {
			err := json.Unmarshal(resp, &time)
			if err != nil {
				fmt.Println("Error with response: ", err)
				p.errorMsg = "There was an error booking the time. " + erb.FriendlyMsg()
				p.showPopup = false
				p.reservSelected = nil
				return
			}
			ctx.SetState("bookRes", time)
			ctx.Navigate("/bookings")
		}
	} else {
		ctx.Navigate("/login")
	}

}
func (p *AvailTimes) onDateChange(ctx app.Context, opts app.Event) {
	p.selectedDate = time.Date(p.selectedDate.Year(), p.selectedDate.Month(), p.selectedDate.Day(), 0, 0, 0, 0, p.selectedDate.Location()).Add(time.Hour * 24)

	ctx.Dispatch(func(ctx app.Context) {
		p.getTeeTimes()
	})
}
func (p *AvailTimes) onDateBack(ctx app.Context, opts app.Event) {
	p.selectedDate = p.selectedDate.Add(time.Hour * -24)
	if p.selectedDate.YearDay() == time.Now().YearDay() {
		p.selectedDate = time.Now()
	}
	ctx.Dispatch(func(ctx app.Context) {
		p.getTeeTimes()
	})
}

func (s *AvailTimes) renderPopup() app.UI {

	return app.Div().Class("reservation-card").Body(
		app.Div().Class("res-hdr").Body(
			app.Div().Text("BigFoot Golf Club"),
		),
		app.Div().Class("res-details").Body(
			app.Div().Body(
				app.Span().Text("Date"),
				app.Span().Text(s.reservSelected.TeeTime.Format("3:04 PM")),
			),
			app.Div().Body(
				app.Span().Text("Tee Time"),
				app.Span().Text(s.reservSelected.TeeTime.Format("3:04 PM")),
			),
			s.renderButtons(),
			app.Div().Body(
				app.Span().Text("Price per Person"),
				app.Span().Text(fmt.Sprintf("$%.2f", s.reservSelected.Price)),
			),
		),
		app.Div().Class("total-rows").Body(
			app.Div().Body(
				app.Span().Text("Grand Total"),
				app.Span().Text(fmt.Sprintf("$%.2f", s.reservSelected.Price*float32(s.players))),
			),
			app.Div().Body(
				app.Button().Text("Book").OnClick(s.onBookSlot),
				app.Button().Text("Cancel").OnClick(func(ctx app.Context, e app.Event) {
					ctx.Dispatch(func(ctx app.Context) {
						s.reservSelected = nil
						s.showPopup = false
					})
				}),
			),
		),
	)

}
func (s *AvailTimes) renderButtons() app.UI {
	playerBtns := []string{"1️⃣", "2️⃣", "3️⃣", "4️⃣"}

	divOut := app.Div().Class("fixedTeeHeader").Body(
		app.Div().Text("Players"),
		app.Range(playerBtns).Slice(func(i int) app.UI {
			clsStr := []string{"fixedTeeBtn"}
			if (i + 1) == s.players {
				clsStr = append(clsStr, "emoji-selected")
			}
			return app.Div().Class(clsStr...).Text(playerBtns[i]).OnClick(func(ctx app.Context, e app.Event) { ctx.Dispatch(func(ctx app.Context) { s.players = i + 1 }) })
		}),
	)
	return divOut
}

func (s *AvailTimes) renderCourseSelector() app.UI {
	return app.Div().Class("course-selector").Body(
		app.Label().For("course-dropdown").Text("Select Course:"),
		app.Select().
			ID("course-dropdown").
			Class("course-dropdown").
			OnChange(s.onCourseChange).
			Body(
				// Default BigFoot option
				app.Option().
					Value("").
					Text("BigFoot Golf Club (Default)").
					Selected(s.selectedCourse == ""),
				// Dynamic course options
				app.Range(s.courses).Slice(func(i int) app.UI {
					course := s.courses[i]
					return app.Option().
						Value(course.ID).
						Text(course.Name).
						Selected(s.selectedCourse == course.ID)
				}),
			),
	)
}

func (p *AvailTimes) onCourseChange(ctx app.Context, e app.Event) {
	courseID := ctx.JSSrc().Get("value").String()
	ctx.Dispatch(func(ctx app.Context) {
		p.selectedCourse = courseID
		p.getTeeTimes()
	})
}
