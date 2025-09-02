package pages

import (
	"bigfoot/golf/app/clients"
	"bigfoot/golf/common/models/anthropic"
	"fmt"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type ChatAgent struct {
	app.Compo
	userMessage   string
	agentMessage  string
	agentResponse anthropic.ChatResponse
	agentRequest  anthropic.ChatRequest
}

func (h *ChatAgent) Render() app.UI {

	return app.Div().
		Class("page").
		Body(
			app.Div().
				Class("page-header").
				Body(
					app.H1().Text("🏌️ Birds Foot Golf"),
					app.P().
						Class("subtitle").
						Text("Book your perfect tee time"),
				),

			app.Div().
				Class("quick-actions").
				Body(
					app.Button().
						Class("action-btn primary").
						Text("Quick Book").
						OnClick(h.onQuickBook),

					app.Button().
						Class("action-btn secondary").
						Text("View Course Info").
						OnClick(h.onCourseInfo),
				),
			app.Div().
				Class("agentchat").
				Body(
					app.Textarea().
						Class("agentchatBox").
						Text(h.agentMessage).
						OnChange(h.ValueTo(&h.agentMessage)),
					app.Br(),
					app.Input().
						Type("text").
						AutoFocus(true).
						OnChange(h.ValueTo(&h.userMessage)),
					app.Br(),
					app.Button().
						Class("action-btn secondary").
						Text("Chat").
						OnClick(h.onChatClick),
				),
			app.Div().
				Class("weather-widget").
				Body(
					app.H3().Text("Today's Weather"),
					app.P().Text("Perfect golfing conditions!"),
					app.Div().
						Class("weather-info").
						Body(
							app.Span().Text("☀️ 75°F"),
							app.Span().Text("💨 5mph winds"),
						),
				),
		)
}

func (h *ChatAgent) onQuickBook(ctx app.Context, e app.Event) {
	ctx.Navigate("/search")
}

func (h *ChatAgent) onCourseInfo(ctx app.Context, e app.Event) {
	// Show course information
	ctx.NewAction("show-course-info")
}

func (h *ChatAgent) onChatClick(ctx app.Context, e app.Event) {
	// Show course information
	if h.userMessage == "" {
		h.agentMessage += "TALK TO ME...."
		ctx.Update()
		return
	}
	fmt.Println("chatting ", h.userMessage)

	h.agentRequest.AddNewMessage(h.userMessage)

	_resp, err := clients.CallAgentProxy(h.agentRequest)
	if err != nil {
		fmt.Println("Error: ", err)
		h.userMessage = ""
		ctx.Update()
		return
	}
	if _resp != nil {
		h.agentResponse = *_resp
		h.agentMessage += _resp.Response
		h.userMessage = ""
		ctx.Update()
	}
}
