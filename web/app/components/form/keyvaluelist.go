package form

import (
	"fmt"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type KeyValuePair struct {
	Key   string
	Value string
}

type KeyValueList struct {
	app.Compo
	Label        string
	Pairs        []KeyValuePair
	showModal    bool
	editingIndex int
	modalKey     string
	modalValue   string
	OnChange     func([]KeyValuePair)
}

func (kv *KeyValueList) Render() app.UI {
	return app.Div().Class("key-value-list").Body(
		app.Label().Text(kv.Label),

		// List of existing pairs
		app.Div().Class("kv-pairs-container").Body(
			app.If(len(kv.Pairs) == 0, func() app.UI {
				return app.P().Class("kv-empty").Text("No items added yet")
			}),
			app.Range(kv.Pairs).Slice(func(i int) app.UI {
				pair := kv.Pairs[i]
				return app.Div().Class("kv-pair-item").Body(
					app.Div().Class("kv-pair-content").Body(
						app.Span().Class("kv-key").Text(fmt.Sprintf("Key: %s", pair.Key)),
						app.Span().Class("kv-value").Text(fmt.Sprintf("Value: %s", pair.Value)),
					),
					app.Div().Class("kv-pair-actions").Body(
						app.Button().
							Class("btn btn-small btn-secondary").
							Text("Edit").
							DataSet("index", fmt.Sprintf("%d", i)).
							OnClick(kv.onEditPair),
						app.Button().
							Class("btn btn-small btn-danger").
							Text("Remove").
							DataSet("index", fmt.Sprintf("%d", i)).
							OnClick(kv.onRemovePair),
					),
				)
			}),
		),

		// Add button
		app.Button().
			Class("btn btn-primary btn-add-kv").
			Text("+ Add Item").
			OnClick(kv.onShowModal),

		// Modal
		app.If(kv.showModal, func() app.UI {
			return kv.renderModal()
		}),
	)
}

func (kv *KeyValueList) renderModal() app.UI {
	return app.Div().Class("modal-overlay").Body(
		app.Div().Class("modal-content").Body(
			app.H3().Text(func() string {
				if kv.editingIndex >= 0 {
					return "Edit Item"
				}
				return "Add New Item"
			}()),

			app.Div().Class("form-group").Body(
				app.Label().For("kvKey").Text("Key:"),
				app.Input().
					Type("text").
					ID("kvKey").
					Value(kv.modalKey).
					Placeholder("Enter key").
					OnInput(kv.onKeyInput),
			),

			app.Div().Class("form-group").Body(
				app.Label().For("kvValue").Text("Value:"),
				app.Input().
					Type("text").
					ID("kvValue").
					Value(kv.modalValue).
					Placeholder("Enter value").
					OnInput(kv.onValueInput),
			),

			app.Div().Class("modal-actions").Body(
				app.Button().
					Class("btn btn-primary").
					Text("Save").
					OnClick(kv.onSavePair),
				app.Button().
					Class("btn btn-secondary").
					Text("Cancel").
					OnClick(kv.onCloseModal),
			),
		),
	).OnClick(kv.onOverlayClick)
}

func (kv *KeyValueList) onShowModal(ctx app.Context, e app.Event) {
	e.PreventDefault()
	ctx.Dispatch(func(ctx app.Context) {
		kv.showModal = true
		kv.editingIndex = -1
		kv.modalKey = ""
		kv.modalValue = ""
	})
}

func (kv *KeyValueList) onCloseModal(ctx app.Context, e app.Event) {
	e.PreventDefault()
	ctx.Dispatch(func(ctx app.Context) {
		kv.showModal = false
		kv.editingIndex = -1
		kv.modalKey = ""
		kv.modalValue = ""
	})
}

func (kv *KeyValueList) onOverlayClick(ctx app.Context, e app.Event) {
	// Close modal if clicking on overlay (not on modal content)
	if ctx.JSSrc().Get("className").String() == "modal-overlay" {
		kv.onCloseModal(ctx, e)
	}
}

func (kv *KeyValueList) onKeyInput(ctx app.Context, e app.Event) {
	kv.modalKey = ctx.JSSrc().Get("value").String()
}

func (kv *KeyValueList) onValueInput(ctx app.Context, e app.Event) {
	kv.modalValue = ctx.JSSrc().Get("value").String()
}

func (kv *KeyValueList) onSavePair(ctx app.Context, e app.Event) {
	e.PreventDefault()

	if kv.modalKey == "" {
		return
	}

	ctx.Dispatch(func(ctx app.Context) {
		if kv.editingIndex >= 0 {
			// Update existing pair
			kv.Pairs[kv.editingIndex] = KeyValuePair{
				Key:   kv.modalKey,
				Value: kv.modalValue,
			}
		} else {
			// Add new pair
			kv.Pairs = append(kv.Pairs, KeyValuePair{
				Key:   kv.modalKey,
				Value: kv.modalValue,
			})
		}

		kv.showModal = false
		kv.editingIndex = -1
		kv.modalKey = ""
		kv.modalValue = ""

		// Notify parent of change
		if kv.OnChange != nil {
			kv.OnChange(kv.Pairs)
		}
	})
}

func (kv *KeyValueList) onEditPair(ctx app.Context, e app.Event) {
	e.PreventDefault()
	indexStr := ctx.JSSrc().Get("dataset").Get("index").String()
	var index int
	fmt.Sscanf(indexStr, "%d", &index)

	if index >= 0 && index < len(kv.Pairs) {
		ctx.Dispatch(func(ctx app.Context) {
			kv.editingIndex = index
			kv.modalKey = kv.Pairs[index].Key
			kv.modalValue = kv.Pairs[index].Value
			kv.showModal = true
		})
	}
}

func (kv *KeyValueList) onRemovePair(ctx app.Context, e app.Event) {
	e.PreventDefault()
	indexStr := ctx.JSSrc().Get("dataset").Get("index").String()
	var index int
	fmt.Sscanf(indexStr, "%d", &index)

	if index >= 0 && index < len(kv.Pairs) {
		ctx.Dispatch(func(ctx app.Context) {
			// Remove the pair
			kv.Pairs = append(kv.Pairs[:index], kv.Pairs[index+1:]...)

			// Notify parent of change
			if kv.OnChange != nil {
				kv.OnChange(kv.Pairs)
			}
		})
	}
}
