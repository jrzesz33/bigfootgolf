// Package handlers provides HTTP request handlers for the golf booking application.
// It registers all API routes and connects HTTP endpoints to business logic in the
// models and controllers packages.
//
// The package is organized into several subrouters:
// - /auth - Authentication endpoints (login, register, OAuth)
// - /api - Authenticated user endpoints (profile, bookings, chat)
// - /papi - Public endpoints (available tee times, course info)
// - /admin - Administrative endpoints (seasons, settings, reservations)
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
