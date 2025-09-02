package handlers

import (
	"bigfoot/golf/common/models/auth"

	"github.com/gorilla/mux"
)

func RegisterAuthRouter(router *mux.Router) {

	authServer := auth.InitAuth()

	// Auth routes
	router.HandleFunc("/register", authServer.HandleRegister).Methods("POST")
	router.HandleFunc("/login", authServer.HandleLogin).Methods("POST")
	router.HandleFunc("/google", authServer.HandleGoogleLogin).Methods("GET")
	router.HandleFunc("/google/callback", authServer.HandleGoogleCallback).Methods("POST")
	router.HandleFunc("/apple", authServer.HandleAppleLogin).Methods("GET")
	router.HandleFunc("/apple/callback", authServer.HandleAppleCallback).Methods("POST")
	router.HandleFunc("/me", authServer.AuthenticateMiddleware(false, authServer.HandleMe)).Methods("GET")
	router.HandleFunc("/refresh", authServer.HandleRefreshToken).Methods("POST")
}
