package server

import (
	"html/template"
	"net/http"
)

type BookNew struct {
	templates *template.Template
}

type BookCreateRequest struct {
	Title        string
	Author       string
	DateFinished string
	Media        string
}

func (action *BookNew) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO - CSRF protection

	bcr := &BookCreateRequest{}

	tmpl := action.templates.Lookup("book_new")
	err := tmpl.Execute(w, map[string]interface{}{"fields": bcr, "errors": nil})
	if err != nil {
		panic(err)
	}
}
