package server

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/jackc/booklog/domain"
	"github.com/jackc/booklog/validate"
	"github.com/jackc/pgx/v4"
	errors "golang.org/x/xerrors"
)

func BookCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(queryExecer)

	username := chi.URLParam(r, "username")
	var userID int64
	err := db.QueryRow(ctx, "select id from users where username=$1", username).Scan(&userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			NotFoundHandler(w, r)
		} else {
			InternalServerErrorHandler(w, r, err)
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

	err = domain.CreateBook(ctx, db, cba)
	if err != nil {
		var verr validate.Errors
		if errors.As(err, &verr) {
			err := RenderBookNew(w, baseViewDataFromRequest(r), cba, verr, username)
			if err != nil {
				InternalServerErrorHandler(w, r, err)
			}
			return
		}

		InternalServerErrorHandler(w, r, err)
		return
	}

	http.Redirect(w, r, BooksPath(username), http.StatusSeeOther)
}
