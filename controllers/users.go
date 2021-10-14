package controllers

import (
	"net/http"

	"github.com/lewismevan/learn-go/views"
)

type Users struct {
	NewView *views.View
}

/*
 * Creates a new Users controller. This function should only
 * be used during setup, since it will panic if a template
 * cannot be parsed correctly.
 */
func NewUsers() *Users {
	return &Users{
		NewView: views.NewView("bootstrap", "views/users/new.gohtml"),
	}
}

func (u *Users) New(w http.ResponseWriter, r *http.Request) {
	if err := u.NewView.Render(w, nil); err != nil {
		panic(err)
	}
}
