package server

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/jackc/booklog/domain"
)

func BookDelete(ctx context.Context, e *Endpoint, w http.ResponseWriter, r *http.Request) {
	bookID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		e.NotFound(w, r)
		return
	}

	var username string
	err = e.DB.QueryRow(ctx, "select username from users join books on users.id=books.user_id where books.id=$1", bookID).Scan(&username)
	if err != nil {
		e.NotFound(w, r)
		return
	}

	dba := domain.DeleteBookArgs{
		ID: bookID,
	}

	err = domain.DeleteBook(ctx, e.DB, dba)
	if err != nil {
		e.InternalServerError(w, r, err)
		return
	}

	http.Redirect(w, r, BooksPath(username), http.StatusSeeOther)
}
