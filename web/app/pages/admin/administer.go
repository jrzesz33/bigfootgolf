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

type AdminSection string

const (
	SectionSeasons AdminSection = "seasons"
	SectionCourses AdminSection = "courses"
)

type Administer struct {
	app.Compo
	components.BaseCo
	seasons           []teetimes.Season
	selectedSeason    *teetimes.Season
	menuDropdown      form.DropDown
	currentSection    AdminSection
	coursesManagement CoursesManagement
}

func (h *Administer) OnMount(ctx app.Context) {
	fmt.Println("Mount Triggered for Admin Page")
	// Set default section
	h.currentSection = SectionSeasons
	h.loadSeasons(ctx)
}

func (h *Administer) OnNav(ctx app.Context) {

}
func (h *Administer) OnDismount() {

}

func (h *Administer) Render() app.UI {
	fmt.Println("Rendering Admin Page")

	return app.Div().Class("admin-container").Body(
		// Section Menu
		app.Div().Class("admin-section-menu").Body(
			app.Button().
				Class(func() string {
					if h.currentSection == SectionSeasons {
						return "admin-menu-btn active"
					}
					return "admin-menu-btn"
				}()).
				Text("Seasons & Tee Times").
				OnClick(h.onSectionClick).
				DataSet("section", string(SectionSeasons)),
			app.Button().
				Class(func() string {
					if h.currentSection == SectionCourses {
						return "admin-menu-btn active"
					}
					return "admin-menu-btn"
				}()).
				Text("Other Courses").
				OnClick(h.onSectionClick).
				DataSet("section", string(SectionCourses)),
		),
		// Content Area
		app.Div().Class("admin-content").Body(
			// Seasons Section
			app.If(h.currentSection == SectionSeasons, func() app.UI {
				return app.Div().Class("admin-section seasons-section").Body(
					app.H2().Text("Manage Seasons & Tee Times"),
					&h.menuDropdown,
					app.If(h.selectedSeason != nil, func() app.UI {
						return app.Div().Body(
							app.P().Text(fmt.Sprintf("%s %d", h.selectedSeason.Name, h.selectedSeason.Year)),
						)
					}),
				)
			}),
			// Courses Section
			app.If(h.currentSection == SectionCourses, func() app.UI {
				return app.Div().Class("admin-section courses-section").Body(
					&h.coursesManagement,
				)
			}),
		),
	)
}

func (h *Administer) onSectionClick(ctx app.Context, e app.Event) {
	e.PreventDefault()
	section := ctx.JSSrc().Get("dataset").Get("section").String()
	ctx.Dispatch(func(ctx app.Context) {
		h.currentSection = AdminSection(section)
		if h.currentSection == SectionCourses {
			h.coursesManagement.LoadCourses(ctx)
		}
	})
}

func (h *Administer) loadSeasons(ctx app.Context) {
	_body, erb := clients.SendPostWithAuth("./admin/seasons", "")
	if erb.Code != 200 {
		ctx.Navigate("./")
		return
	}

	err := json.Unmarshal(_body, &h.seasons)
	if err != nil {
		fmt.Println(err)
		return
	}
	sort.Slice(h.seasons, func(i, j int) bool {
		return h.seasons[i].BeginDate.Before(h.seasons[j].BeginDate)
	})
	h.menuDropdown.MenuMap = nil
	for _, seas := range h.seasons {
		_item := make(map[string]string)
		_item["value"] = seas.ID
		_item["name"] = fmt.Sprintf("%d %s", seas.Year, seas.Name)
		h.menuDropdown.MenuMap = append(h.menuDropdown.MenuMap, _item)
	}
	h.menuDropdown.MenuSelect = h.onSeasonClick
}

func (h *Administer) onSeasonClick(val string) {
	for _, seas := range h.seasons {
		if seas.ID == val {
			h.selectedSeason = &seas
			return
		}
	}
	fmt.Println("season value-", val)
}
