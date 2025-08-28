//go:build js && wasm
// +build js,wasm

package main

import (
	"birdsfoot/app/app/routes"
	"birdsfoot/app/controllers"
	"birdsfoot/app/handlers"
	"birdsfoot/app/handlers/sessionmgr"
	"birdsfoot/app/models/db"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

func main() {

	routes.RegisterRoutes()

	//initialize the tokenManager
	//state.GetAppState()

	// Run when on the client side
	app.RunWhenOnBrowser()

	wasmHandler := &app.Handler{
		Name:        "Golf Booking App",
		Description: "A mobile-friendly golf tee time booking app",
		//Icon: app.Icon{
		//	Default: "/web/icon-192.png",
		//},
		Keywords: []string{
			"golf",
			"booking",
			"tee times",
		},
		LoadingLabel: "Loading Golf App...",
		Styles: []string{
			"/web/app.css",
			"/web/app_add.css",
			"/web/nav.css",
			"/web/agent.css",
		},
		Scripts: []string{
			// Add any external scripts here
		},
	}

	//*************** Initialize Server Side Systems ********************************
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		log.Fatalf("Error loading location: %v", err)
	}
	time.Local = loc // Set the global timezone
	fmt.Println("Application timezone set to:", time.Local.String())

	//Initialize the Database
	ctx := context.Background()
	db.InitDB(ctx)

	// Create a new router
	r := mux.NewRouter()

	// Apply global middleware
	r.Use(loggingMiddleware)
	// Create API subrouter
	api := r.PathPrefix("/api").Subrouter()
	handlers.RegisterAPIRoutes(api)
	// Create Public subrouter
	papi := r.PathPrefix("/papi").Subrouter()
	handlers.RegisterPublicRoutes(papi)
	// Create API subrouter
	authRouter := r.PathPrefix("/auth").Subrouter()
	handlers.RegisterAuthRouter(authRouter)
	//Create Admin subrouter
	adminRouter := r.PathPrefix("/admin").Subrouter()
	handlers.RegisterAdminRoutes(adminRouter)

	//initialize session Manager
	sessionmgr.NewSessionMgr()

	// Serve static files (optional)
	r.PathPrefix("/web/").Handler(http.StripPrefix("/web/", http.FileServer(http.Dir("./web"))))

	// This ensures go-app's client-side routing takes over for PWA navigation
	r.PathPrefix("/").Handler(wasmHandler)

	if os.Getenv("MODE") == "dev" {
		controllers.SetupDevEnvironment()
	}
	// Start server
	port := ":8001"
	fmt.Printf("Server starting on port %s\n", port)

	log.Fatal(http.ListenAndServe(port, r))
}

// Middleware for logging requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.RequestURI, time.Since(start))
	})
}
