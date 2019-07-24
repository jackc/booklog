package server

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/jackc/booklog/domain"
)

func BookDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(queryExecer)

	bookID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		NotFoundHandler(w, r)
		return
	}

	var username string
	err = db.QueryRow(ctx, "select username from users join books on users.id=books.user_id where books.id=$1", bookID).Scan(&username)
	if err != nil {
		NotFoundHandler(w, r)
		return
	}

	dba := domain.DeleteBookArgs{
		ID: bookID,
	}

	err = domain.DeleteBook(ctx, db, dba)
	if err != nil {
		InternalServerErrorHandler(w, r, err)
		return
	}

	http.Redirect(w, r, BooksPath(username), http.StatusSeeOther)
}
