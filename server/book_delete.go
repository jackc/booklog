package server

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/jackc/booklog/domain"
)

type BookDelete struct {
}

func (action *BookDelete) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	dba := domain.DeleteBookArgs{
		ID: bookID,
	}

	err = domain.DeleteBook(ctx, db, dba)
	if err != nil {
		panic(err)
	}

	http.Redirect(w, r, BooksPath(username), http.StatusSeeOther)

}
