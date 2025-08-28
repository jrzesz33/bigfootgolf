package handlers

import (
	"birdsfoot/app/handlers/transactions"
	"birdsfoot/app/models/auth"

	"github.com/gorilla/mux"
)

func RegisterAPIRoutes(router *mux.Router) {

	authServer := auth.InitAuth()
	// Authenticated routes
	router.HandleFunc("/chat", authServer.AuthenticateMiddleware(false, GetChatHandler)).Methods("POST")
	router.HandleFunc("/userupdate", authServer.AuthenticateMiddleware(false, transactions.SaveUserHandler)).Methods("POST")
	router.HandleFunc("/verifyreq", authServer.AuthenticateMiddleware(false, transactions.SendEmailCodeHandler)).Methods("POST")
	router.HandleFunc("/verifyemailcode", authServer.AuthenticateMiddleware(false, transactions.VerifyCodeHandler)).Methods("POST")
	router.HandleFunc("/resetapw", authServer.AuthenticateMiddleware(false, transactions.UpdatePW)).Methods("POST")

}
