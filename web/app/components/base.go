package components

import (
	"bigfoot/golf/common/models/account"
	"bigfoot/golf/common/models/auth"
	"fmt"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type BaseCo struct {
	User     account.User
	authResp auth.AuthResponse
	SecLevel auth.AuthLevel
}

func (b *BaseCo) GetFromState(ctx app.Context) {
	ctx.GetState(StateKey, &b.authResp)
	if b.authResp.Token == "" {
		fmt.Println("Base Has No Local Storate")
		return
	}

	b.User = b.authResp.User
	b.SecLevel = b.authResp.AuthLevel
	ctx.Update()

}
func (b *BaseCo) ObserveFromState(ctx app.Context) {
	//fmt.Println("getting state ", StateKey)
	ctx.ObserveState(StateKey, &b.authResp).
		OnChange(func() {
			fmt.Println("updating BASE STATE for User ", b.authResp.User.FirstName, b.authResp.User.LastName)
			b.User = b.authResp.User
			b.SecLevel = b.authResp.AuthLevel
			ctx.Update()
		})
}
