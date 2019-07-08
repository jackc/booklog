package server

import (
	"net/http"

	"github.com/gorilla/csrf"
)

type BookNew struct {
}

type BookCreateRequest struct {
	Title        string
	Author       string
	DateFinished string
	Media        string
}

func (action *BookNew) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	bcr := &BookCreateRequest{}
	err := RenderBookNew(w, csrf.TemplateField(r), bcr, map[string]string{})
	if err != nil {
		panic(err)
	}
}
