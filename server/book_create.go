package server

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/jackc/booklog/domain"
	"github.com/jackc/booklog/validate"
	"github.com/jackc/pgx/v4"
	errors "golang.org/x/xerrors"
)

func BookCreate(ctx context.Context, e *Endpoint, w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	var userID int64
	err := e.DB.QueryRow(ctx, "select id from users where username=$1", username).Scan(&userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			e.NotFound(w, r)
		} else {
			e.InternalServerError(w, r, err)
		}
		return
	}

	cba := domain.CreateBookArgs{
		ReaderID:     userID,
		Title:        r.FormValue("title"),
		Author:       r.FormValue("author"),
		DateFinished: r.FormValue("dateFinished"),
		Media:        r.FormValue("media"),
	}

	err = domain.CreateBook(ctx, e.DB, cba)
	if err != nil {
		var verr validate.Errors
		if errors.As(err, &verr) {
			err := RenderBookNew(w, baseViewDataFromRequest(r), cba, verr, username)
			if err != nil {
				e.InternalServerError(w, r, err)
			}
			return
		}

		e.InternalServerError(w, r, err)
		return
	}

	http.Redirect(w, r, BooksPath(username), http.StatusSeeOther)
}
