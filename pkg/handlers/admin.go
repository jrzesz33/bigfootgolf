package handlers

import (
	"bigfoot/golf/common/handlers/admin"
	"bigfoot/golf/common/models/auth"

	"github.com/gorilla/mux"
)

func RegisterAdminRoutes(router *mux.Router) {

	authServer := auth.InitAuth()
	// Authenticated routes
	router.HandleFunc("/seasons", authServer.AuthenticateMiddleware(true, admin.GetSeasons)).Methods("POST")

	// Course management routes
	router.HandleFunc("/courses/list", authServer.AuthenticateMiddleware(true, admin.GetCourses)).Methods("POST")
	router.HandleFunc("/courses/save", authServer.AuthenticateMiddleware(true, admin.SaveCourse)).Methods("POST")
	router.HandleFunc("/courses/toggle", authServer.AuthenticateMiddleware(true, admin.ToggleCourseActive)).Methods("POST")

}
