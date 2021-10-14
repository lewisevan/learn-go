package controllers

import (
	"fmt"
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

/*
 * GET /signup
 * Renders the signup page for a new user.
 */
func (u *Users) New(w http.ResponseWriter, r *http.Request) {
	if err := u.NewView.Render(w, nil); err != nil {
		panic(err)
	}
}

/*
 * POST /signup
 * Processes the signup form data and creates a new user account.
 */
func (u *Users) Create(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "This is a temporary response.")
}
