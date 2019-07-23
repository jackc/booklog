package server

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/jackc/booklog/domain"
	"github.com/jackc/pgx/v4"
	errors "golang.org/x/xerrors"
)

func BookEdit(ctx context.Context, e *Endpoint, w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	bookID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		e.NotFound(w, r)
		return
	}

	uba := domain.UpdateBookArgs{}
	err = e.DB.QueryRow(ctx, "select title, author, finish_date::text, media from books where id=$1", bookID).Scan(&uba.Title, &uba.Author, &uba.DateFinished, &uba.Media)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			e.NotFound(w, r)
		} else {
			e.InternalServerError(w, r, err)
		}
		return
	}

	err = RenderBookEdit(w, baseViewDataFromRequest(r), bookID, uba, nil, username)
	if err != nil {
		e.InternalServerError(w, r, err)
		return
	}
}
