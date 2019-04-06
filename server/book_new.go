package server

import (
	"html/template"
	"net/http"

	"github.com/gorilla/csrf"
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
	bcr := &BookCreateRequest{}

	tmpl := action.templates.Lookup("book_new")
	err := tmpl.Execute(w, map[string]interface{}{"fields": bcr, "errors": map[string]string{}, csrf.TemplateTag: csrf.TemplateField(r)})
	if err != nil {
		panic(err)
	}
}
