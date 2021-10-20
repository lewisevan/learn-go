package controllers

import (
	"fmt"
	"net/http"

	"github.com/lewismevan/learn-go/models"
	"github.com/lewismevan/learn-go/views"
)

type Users struct {
	NewView *views.View
	us      *models.UserService
}

/*
 * Creates a new Users controller. This function should only
 * be used during setup, since it will panic if a template
 * cannot be parsed correctly.
 */
func NewUsers(us *models.UserService) *Users {
	return &Users{
		NewView: views.NewView("bootstrap", "users/new"),
		us:      us,
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

type SignupForm struct {
	Name     string `schema:"name"`
	Email    string `schema:"email"`
	Password string `schema:"password"`
}

/*
 * POST /signup
 * Processes the signup form data and creates a new user account.
 */
func (u *Users) Create(w http.ResponseWriter, r *http.Request) {
	var form SignupForm
	if err := parseForm(r, &form); err != nil {
		panic(err)
	}

	user := models.User{
		Name:     form.Name,
		Email:    form.Email,
		Password: form.Password,
	}

	if err := u.us.Create(&user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, user)
}
