package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/lewismevan/learn-go/controllers"
)

func main() {
	// Controllers
	staticC := controllers.NewStatic()
	usersC := controllers.NewUsers()

	// Create server
	r := mux.NewRouter()

	// Handle Routes
	r.Handle("/", staticC.Home).Methods("GET")
	r.Handle("/contact", staticC.Contact).Methods("GET")

	r.HandleFunc("/signup", usersC.New).Methods("GET")
	r.HandleFunc("/signup", usersC.Create).Methods("POST")

	// Start server
	http.ListenAndServe("localhost:3000", r)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
