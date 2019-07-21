package server

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/gorilla/csrf"
	"github.com/jackc/booklog/domain"
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
	err = db.QueryRow(ctx, "select username from login_account join finished_book on login_account.id=finished_book.reader_id where finished_book.id=$1", bookID).Scan(&username)
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
		// TODO - if errors is not a map this fails
		err := RenderBookEdit(w, csrf.TemplateField(r), bookID, uba, err, username)
		if err != nil {
			panic(err)
		}
		return
	}

	http.Redirect(w, r, BooksPath(username), http.StatusSeeOther)

}
