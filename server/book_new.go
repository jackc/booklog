package server

import (
	"net/http"

	"github.com/go-chi/chi"
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
	err := RenderBookNew(w, baseViewDataFromRequest(r), createBookArgs, nil, username)
	if err != nil {
		panic(err)
	}
}
