package handlers

import (
	"birdsfoot/app/handlers/admin"
	"birdsfoot/app/models/auth"

	"github.com/gorilla/mux"
)

func RegisterAdminRoutes(router *mux.Router) {

	authServer := auth.InitAuth()
	// Authenticated routes
	router.HandleFunc("/seasons", authServer.AuthenticateMiddleware(true, admin.GetSeasons)).Methods("POST")

}
