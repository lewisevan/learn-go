package controllers

import (
	"github.com/lewismevan/learn-go/views"
)

type Static struct {
	Home    *views.View
	Contact *views.View
}

/*
 * Creates a new Static controller for serving static pages.
 */
func NewStatic() *Static {
	return &Static{
		Home:    views.NewView("bootstrap", "views/static/home.gohtml"),
		Contact: views.NewView("bootstrap", "views/static/contact.gohtml"),
	}
}
