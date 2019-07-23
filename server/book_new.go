package server

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/jackc/booklog/domain"
)

func BookNew(ctx context.Context, e *Endpoint, w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")

	var createBookArgs domain.CreateBookArgs
	err := RenderBookNew(w, baseViewDataFromRequest(r), createBookArgs, nil, username)
	if err != nil {
		e.InternalServerError(w, r, err)
		return
	}
}
