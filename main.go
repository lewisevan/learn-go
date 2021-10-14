package main

import (
	"fmt"
	"net/http"
	"text/template"

	"github.com/gorilla/mux"
)

var homeTemplate *template.Template

func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	if err := homeTemplate.Execute(w, nil); err != nil {
		panic(err)
	}
}

var contactTemplate *template.Template

func contact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	if err := contactTemplate.Execute(w, nil); err != nil {
		panic(err)
	}
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "<p>Sorry, but we could not find the page you were looking for.</p>")
}

func main() {
	var err error
	homeTemplate, err = template.ParseFiles(
		"views/home.gohtml",
		"views/layouts/footer.gohtml",
	)
	if err != nil {
		panic(err)
	}

	contactTemplate, err = template.ParseFiles(
		"views/contact.gohtml",
		"views/layouts/footer.gohtml",
	)
	if err != nil {
		panic(err)
	}

	r := mux.NewRouter()
	r.NotFoundHandler = http.HandlerFunc(notFound)
	r.HandleFunc("/", home)
	r.HandleFunc("/contact", contact)
	http.ListenAndServe("localhost:3000", r)
}
