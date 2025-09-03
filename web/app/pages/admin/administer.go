package admin

import (
	"bigfoot/golf/common/models/teetimes"
	"bigfoot/golf/web/app/clients"
	"bigfoot/golf/web/app/components"
	"bigfoot/golf/web/app/components/form"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Administer struct {
	app.Compo
	components.BaseCo
	seasons        []teetimes.Season
	selectedSeason *teetimes.Season
	menuDropdown   form.DropDown
}

func (h *Administer) OnMount(ctx app.Context) {
	fmt.Println("Mount Triggered for Admin Page")
	//Get the List of Seasons
	_body, erb := clients.SendPostWithAuth("./admin/seasons", "")
	if erb.Code != 200 {
		ctx.Navigate("./")
	}

	err := json.Unmarshal(_body, &h.seasons)
	if err != nil {
		fmt.Println(err)
		return
	}
	sort.Slice(h.seasons, func(i, j int) bool {
		return h.seasons[i].BeginDate.Before(h.seasons[j].BeginDate)
	})
	for _, seas := range h.seasons {
		_item := make(map[string]string)
		_item["value"] = seas.ID
		_item["name"] = fmt.Sprintf("%d %s", seas.Year, seas.Name)
		h.menuDropdown.MenuMap = append(h.menuDropdown.MenuMap, _item)
	}
	h.menuDropdown.MenuSelect = h.onSeasonClick
}

func (h *Administer) OnNav(ctx app.Context) {

}
func (h *Administer) OnDismount() {

}

func (h *Administer) Render() app.UI {
	fmt.Println("Rendering Admin Page")

	return app.Div().Body(
		&h.menuDropdown,
		app.If(h.selectedSeason != nil, func() app.UI {
			return app.Div().Body(
				app.P().Text(fmt.Sprintf("%s %d", h.selectedSeason.Name, h.selectedSeason.Year)),
			)
		}),
	)
}
func (h *Administer) onSeasonClick(val string) {
	//fmt.Println("Value of Click", e.Value)
	for _, seas := range h.seasons {
		if seas.ID == val {
			h.selectedSeason = &seas
			return
		}
	}
	fmt.Println("season value-", val)
}
