package server

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
)

type BookDelete struct {
}

type BookDeleteRequest struct {
	ID string
}

func deleteBook(ctx context.Context, db queryExecer, bcr *BookDeleteRequest) error {
	_, err := db.Exec(ctx, "delete from finished_book where id=$1", bcr.ID)
	return err
}

func (action *BookDelete) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(queryExecer)
	bookID := chi.URLParam(r, "id")

	var username string
	err := db.QueryRow(ctx, "select username from login_account join finished_book on login_account.id=finished_book.reader_id where finished_book.id=$1", bookID).Scan(&username)
	if err != nil {
		panic(err)
	}

	bcr := &BookDeleteRequest{}
	bcr.ID = bookID

	err = deleteBook(ctx, db, bcr)
	if err != nil {
		panic(err)
	}

	http.Redirect(w, r, BooksPath(username), http.StatusSeeOther)

}
