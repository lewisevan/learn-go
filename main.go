package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/lewismevan/learn-go/controllers"
	"github.com/lewismevan/learn-go/views"
)

var (
	homeView    *views.View
	contactView *views.View
)

func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	must(homeView.Render(w, nil))
}

func contact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	must(contactView.Render(w, nil))
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "<p>Sorry, but we could not find the page you were looking for.</p>")
}

func main() {
	// Static Views
	homeView = views.NewView("bootstrap", "views/home.gohtml")
	contactView = views.NewView("bootstrap", "views/contact.gohtml")

	// Controllers
	usersC := controllers.NewUsers()

	// Create server
	r := mux.NewRouter()
	r.NotFoundHandler = http.HandlerFunc(notFound)
	r.HandleFunc("/", home)
	r.HandleFunc("/contact", contact)
	r.HandleFunc("/signup", usersC.New)
	http.ListenAndServe("localhost:3000", r)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
