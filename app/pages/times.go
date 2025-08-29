package pages

import (
	"birdsfoot/app/app/clients"
	"birdsfoot/app/models/teetimes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"
	_ "time/tzdata"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type AvailTimes struct {
	app.Compo
	selectedDate time.Time
	//playerCount  int
	timeSlots []teetimes.ReservedDay
}

func (p *AvailTimes) OnMount(ctx app.Context) {
	p.selectedDate = time.Now().Local()
	fmt.Println("Selected Date: ", p.selectedDate)
	//call the web service
	p.getTeeTimes()

}

func (p *AvailTimes) getTeeTimes() {
	//call the web service
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
		//return &_usr, err
	} else {
		fmt.Println(err)
	}
}

func (s *AvailTimes) Render() app.UI {
	if len(s.timeSlots) == 0 {
		return app.Div().Text("No Tee Times for " + s.selectedDate.Format("01/02/2006"))
	}

	x := 0

	sort.Slice(s.timeSlots[x].Times, func(i, j int) bool {
		return s.timeSlots[x].Times[i].TeeTime.Before(s.timeSlots[x].Times[j].TeeTime)
	})
	fmt.Println("days and slots ", len(s.timeSlots), " - ", len(s.timeSlots[0].Times))
	_obj := app.Div().Body(
		app.Div().Class("fixedTeeHeader").Body(
			app.If(s.selectedDate.After(time.Now().Truncate(24*time.Hour).Local().Add(time.Hour*24)), func() app.UI {
				return app.Button().Text("Back").OnClick(s.onDateBack)
			}),
			app.H2().Text(s.selectedDate.Format("Jan 2 Mon")),
			app.Button().Text("Next").OnClick(s.onDateChange)),
		app.Div().
			Class("time-slots").
			Body(
				app.Range(s.timeSlots[x].Times).Slice(func(i int) app.UI {
					slot := s.timeSlots[x].Times[i]
					_p := len(slot.Players)

					if _p < 4 {
						_open := strconv.Itoa(4 - _p)
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
									Text("0.00"),
								app.Button().
									Class("btn secondary").
									Text("Book").
									OnClick(s.onBookSlot),
							)
					}
					return nil
				}),
			),
	)
	return _obj
}

func (p *AvailTimes) onBookSlot(ctx app.Context, opts app.Event) {
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
