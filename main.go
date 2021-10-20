package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/lewismevan/learn-go/controllers"
	"github.com/lewismevan/learn-go/models"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "lenslocked_dev"
)

func main() {
	// Model services
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	us, err := models.NewUserService(psqlInfo)
	must(err)
	defer us.Close()
	us.AutoMigrate()

	// Controllers
	staticC := controllers.NewStatic()
	usersC := controllers.NewUsers(us)

	// Create server
	r := mux.NewRouter()

	// Handle Routes
	r.Handle("/", staticC.Home).Methods("GET")
	r.Handle("/contact", staticC.Contact).Methods("GET")

	r.Handle("/signup", usersC.NewView).Methods("GET")
	r.HandleFunc("/signup", usersC.Create).Methods("POST")

	r.Handle("/login", usersC.LoginView).Methods("GET")
	r.HandleFunc("/login", usersC.Login).Methods("POST")

	// Start server
	http.ListenAndServe("localhost:3000", r)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
