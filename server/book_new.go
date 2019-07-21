package server

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/gorilla/csrf"
	"github.com/jackc/booklog/domain"
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
	username := chi.URLParam(r, "username")

	var createBookArgs domain.CreateBookArgs
	err := RenderBookNew(w, csrf.TemplateField(r), createBookArgs, map[string]string{}, username)
	if err != nil {
		panic(err)
	}
}
