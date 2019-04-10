package server

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/jackc/pgx"
	"github.com/spf13/viper"
)

type BookDelete struct {
}

type BookDeleteRequest struct {
	ID string
}

func deleteBook(bcr *BookDeleteRequest) error {
	conn, err := pgx.Connect(context.Background(), viper.GetString("database_uri"))
	if err != nil {
		panic(err)
	}
	defer conn.Close(context.Background())

	_, err = conn.Exec(context.Background(), "delete from book where id=$1", bcr.ID)
	if err != nil {
		return err
	}

	return nil
}

func (action *BookDelete) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	bcr := &BookDeleteRequest{}
	bcr.ID = chi.URLParam(r, "id")

	err := deleteBook(bcr)
	if err != nil {
		panic(err)
	}

	http.Redirect(w, r, BooksPath(), http.StatusSeeOther)

}
