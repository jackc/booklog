package server

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/jackc/booklog/domain"
	"github.com/jackc/pgx/v4"
	errors "golang.org/x/xerrors"
)

func BookEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(queryExecer)

	username := chi.URLParam(r, "username")
	bookID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		NotFoundHandler(w, r)
		return
	}

	uba := domain.UpdateBookArgs{}
	err = db.QueryRow(ctx, "select title, author, finish_date::text, media from books where id=$1", bookID).Scan(&uba.Title, &uba.Author, &uba.DateFinished, &uba.Media)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			NotFoundHandler(w, r)
		} else {
			InternalServerErrorHandler(w, r, err)
		}
		return
	}

	err = RenderBookEdit(w, baseViewDataFromRequest(r), bookID, uba, nil, username)
	if err != nil {
		InternalServerErrorHandler(w, r, err)
		return
	}
}
