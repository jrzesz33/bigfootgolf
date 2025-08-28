package components

import (
	"birdsfoot/app/models/account"
	"birdsfoot/app/models/auth"
	"fmt"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type BaseCo struct {
	User     account.User
	authResp auth.AuthResponse
	SecLevel auth.AuthLevel
}

func (b *BaseCo) GetFromState(ctx app.Context) {
	ctx.GetState(STATE_KEY, &b.authResp)
	if b.authResp.Token == "" {
		fmt.Println("Base Has No Local Storate")
		return
	}

	b.User = b.authResp.User
	b.SecLevel = b.authResp.AuthLevel
	ctx.Update()

}
func (b *BaseCo) ObserveFromState(ctx app.Context) {
	//fmt.Println("getting state ", STATE_KEY)
	ctx.ObserveState(STATE_KEY, &b.authResp).
		OnChange(func() {
			fmt.Println("updating BASE STATE for User ", b.authResp.User.FirstName, b.authResp.User.LastName)
			b.User = b.authResp.User
			b.SecLevel = b.authResp.AuthLevel
			ctx.Update()
		})
}
