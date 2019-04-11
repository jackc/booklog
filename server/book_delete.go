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
	_, err := db.Exec(ctx, "delete from book where id=$1", bcr.ID)
	return err
}

func (action *BookDelete) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(queryExecer)

	bcr := &BookDeleteRequest{}
	bcr.ID = chi.URLParam(r, "id")

	err := deleteBook(ctx, db, bcr)
	if err != nil {
		panic(err)
	}

	http.Redirect(w, r, BooksPath(), http.StatusSeeOther)

}
