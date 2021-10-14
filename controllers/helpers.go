package controllers

import (
	"net/http"

	"github.com/gorilla/schema"
)

/*
 * Parses form data from an HTTP Request into a specified struct type.
 * The parameter must be a pointer to the struct type in order to
 * return the updated data. Uses gorilla/schema library.
 */
func parseForm(r *http.Request, dst interface{}) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	dec := schema.NewDecoder()

	if err := dec.Decode(dst, r.PostForm); err != nil {
		return err
	}

	return nil
}
