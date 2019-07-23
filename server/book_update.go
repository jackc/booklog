package server

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/jackc/booklog/domain"
	"github.com/jackc/booklog/validate"
	errors "golang.org/x/xerrors"
)

type BookUpdate struct {
}

func (action *BookUpdate) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(queryExecer)
	bookID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		panic(err)
	}

	var username string
	err = db.QueryRow(ctx, "select username from users join books on users.id=books.user_id where books.id=$1", bookID).Scan(&username)
	if err != nil {
		panic(err)
	}

	uba := domain.UpdateBookArgs{
		ID:           bookID,
		Title:        r.FormValue("title"),
		Author:       r.FormValue("author"),
		DateFinished: r.FormValue("dateFinished"),
		Media:        r.FormValue("media"),
	}

	err = domain.UpdateBook(ctx, db, uba)
	if err != nil {
		var verr validate.Errors
		if errors.As(err, &verr) {
			err := RenderBookEdit(w, baseViewDataFromRequest(r), bookID, uba, verr, username)
			if err != nil {
				panic(err)
			}

			return
		}

		panic(err)
	}

	http.Redirect(w, r, BooksPath(username), http.StatusSeeOther)

}
