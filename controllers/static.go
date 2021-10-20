package controllers

import (
	"github.com/lewismevan/learn-go/views"
)

type Static struct {
	Home    *views.View
	Contact *views.View
}

// Creates a new Static controller for serving static pages.
func NewStatic() *Static {
	return &Static{
		Home:    views.NewView("bootstrap", "static/home"),
		Contact: views.NewView("bootstrap", "static/contact"),
	}
}
