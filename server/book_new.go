package server

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/jackc/booklog/domain"
)

func BookNew(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")

	var createBookArgs domain.CreateBookArgs
	err := RenderBookNew(w, baseViewDataFromRequest(r), createBookArgs, nil, username)
	if err != nil {
		InternalServerErrorHandler(w, r, err)
		return
	}
}
