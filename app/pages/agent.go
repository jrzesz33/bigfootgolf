package pages

import (
	"birdsfoot/app/app/clients"
	"birdsfoot/app/models/anthropic"
	"time"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Agent struct {
	app.Compo
	userInput      string
	messages       []ChatMessage
	isLoading      bool
	conversationID string
	currentMessage anthropic.ChatRequest
}

type ChatMessage struct {
	Content   string
	IsUser    bool
	Timestamp time.Time
}

func (a *Agent) OnMount(ctx app.Context) {
	// Initialize with welcome message from BigfootAI
	welcomeMsg := ChatMessage{
		Content:   "What can I help you with today?",
		IsUser:    false,
		Timestamp: time.Now(),
	}
	a.messages = append(a.messages, welcomeMsg)
	a.currentMessage = anthropic.ChatRequest{
		ConversationHist: []anthropic.Message{},
		MaxTokens:        4096,
		Temperature:      0.7,
	}
}

func (a *Agent) Render() app.UI {
	return app.Div().
		Class("agent-container").
		Body(
			app.Main().
				Class("chat-main").
				Body(
					a.renderChatContainer(),
					a.renderInputArea(),
				),
		)
}

func (a *Agent) renderChatContainer() app.UI {
	return app.Div().
		Class("chat-container").
		Body(
			app.Div().
				Class("messages").
				ID("messages").
				Body(
					app.Range(a.messages).Slice(func(i int) app.UI {
						msg := a.messages[i]
						return a.renderMessage(msg)
					}),
					app.If(a.isLoading, func() app.UI {
						return app.Div().
							Class("message ai-message loading").
							Body(
								app.Div().
									Class("message-header").
									Body(
										app.Strong().
											Class("sender").
											Text("BigfootAI: "),
									),
								app.Div().
									Class("message-content").
									Body(
										app.Div().
											Class("typing-indicator").
											Body(
												app.Span().Text("●"),
												app.Span().Text("●"),
												app.Span().Text("●"),
											),
									),
							)
					}),
				),
		)
}

func (a *Agent) renderMessage(msg ChatMessage) app.UI {
	messageClass := "message ai-message"
	sender := "BigfootAI: "

	if msg.IsUser {
		messageClass = "message user-message"
		sender = "John: " // This could be dynamic based on user data
	}

	return app.Div().
		Class(messageClass).
		Body(
			app.Div().
				Class("message-header").
				Body(
					app.Strong().
						Class("sender").
						Text(sender),
				),
			app.Div().
				Class("message-content").
				Text(msg.Content),
			a.renderQuickActions(msg),
		)
}

func (a *Agent) renderQuickActions(msg ChatMessage) app.UI {
	// Only show quick actions for AI messages that suggest booking options
	if msg.IsUser || a.isLoading {
		return app.Text("")
	}

	// Check if this is a response about tee times
	if len(msg.Content) > 50 && (containsTimeSlot(msg.Content)) {
		return app.Div().
			Class("quick-actions").
			Body(
				app.Button().
					Class("quick-action-btn").
					Text("10:20 AM").
					OnClick(a.onQuickAction("10:20 AM")),
				app.Button().
					Class("quick-action-btn").
					Text("11:04 AM").
					OnClick(a.onQuickAction("11:04 AM")),
				app.Button().
					Class("quick-action-btn").
					Text("11:38 AM").
					OnClick(a.onQuickAction("11:38 AM")),
				app.Button().
					Class("quick-action-btn").
					Text("Check afternoon").
					OnClick(a.onQuickAction("Check afternoon")),
			)
	}

	return app.Text("")
}

func (a *Agent) renderInputArea() app.UI {
	return app.Div().
		Class("input-area").
		Body(
			app.Div().
				Class("input-container").
				Body(
					app.Textarea().
						Class("message-input").
						Placeholder("Type your message...").
						Rows(1).
						Text(a.userInput).
						AutoFocus(true).
						OnChange(a.ValueTo(&a.userInput)).
						OnKeyDown(a.onInputKeyDown),
					app.Button().
						Class("send-button").
						Disabled(a.isLoading || a.userInput == "").
						Text("Send").
						OnClick(a.onSendClick),
				),
		)
}

func (a *Agent) onInputKeyDown(ctx app.Context, e app.Event) {
	if e.Get("key").String() == "Enter" && !e.Get("shiftKey").Bool() {
		e.PreventDefault()
		if !a.isLoading && a.userInput != "" {
			a.sendMessage(ctx)
		}
	}
}

func (a *Agent) onSendClick(ctx app.Context, e app.Event) {
	if !a.isLoading && a.userInput != "" {
		a.sendMessage(ctx)
	}
}

func (a *Agent) onQuickAction(action string) func(ctx app.Context, e app.Event) {
	return func(ctx app.Context, e app.Event) {
		a.userInput = action
		a.sendMessage(ctx)
	}
}

func (a *Agent) sendMessage(ctx app.Context) {
	if a.userInput == "" {
		return
	}

	// Add user message to chat
	userMsg := ChatMessage{
		Content:   a.userInput,
		IsUser:    true,
		Timestamp: time.Now(),
	}
	a.messages = append(a.messages, userMsg)

	// Add to conversation history
	a.currentMessage.AddNewMessage(a.userInput)

	// Clear input and show loading
	a.userInput = ""
	a.isLoading = true
	ctx.Update()

	// Make API call
	go func() {
		ctx.Async(func() {
			resp, err := clients.CallAgentProxy(a.currentMessage)

			ctx.Dispatch(func(ctx app.Context) {
				a.isLoading = false

				if err != nil {
					errorMsg := ChatMessage{
						Content:   "Sorry, I'm having trouble connecting right now. Please try again.",
						IsUser:    false,
						Timestamp: time.Now(),
					}
					a.messages = append(a.messages, errorMsg)
				} else if resp != nil {
					aiMsg := ChatMessage{
						Content:   resp.Response,
						IsUser:    false,
						Timestamp: time.Now(),
					}
					a.messages = append(a.messages, aiMsg)
					a.conversationID = resp.ConversationID
					a.currentMessage.ConversationHist = resp.ConversationHist
				}

				ctx.Update()

				// Scroll to bottom after update
				ctx.After(100*time.Millisecond, func(ctx app.Context) {
					// Auto-scroll will happen due to CSS scroll-behavior: smooth
				})
			})
		})
	}()
}

// Helper function to check if content contains time slots
func containsTimeSlot(content string) bool {
	timeKeywords := []string{"AM", "PM", "tee time", "available", "book", "options"}
	for _, keyword := range timeKeywords {
		if len(content) > 0 {
			// Simple check - in real implementation you'd use proper string searching
			for i := 0; i < len(content)-len(keyword)+1; i++ {
				match := true
				for j := 0; j < len(keyword); j++ {
					if content[i+j] != keyword[j] && content[i+j] != keyword[j]+32 && content[i+j] != keyword[j]-32 {
						match = false
						break
					}
				}
				if match {
					return true
				}
			}
		}
	}
	return false
}
