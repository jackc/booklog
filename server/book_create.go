package server

import (
	"context"
	"net/http"

	"github.com/jackc/pgconn"
	"github.com/spf13/viper"
)

type BookCreate struct {
}

type BookCreateRequest struct {
	Title        string
	Author       string
	DateFinished string
	Media        string
}

func createBook(bcr *BookCreateRequest) error {
	conn, err := pgconn.Connect(context.Background(), viper.GetString("database_uri"))
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	result := conn.ExecParams(context.Background(), "insert into book(title, author, date_finished, media) values($1, $2, $3, $4)", [][]byte{[]byte(bcr.Title), []byte(bcr.Author), []byte(bcr.DateFinished), []byte(bcr.Media)}, nil, nil, nil).Read()
	if result.Err != nil {
		return err
	}

	return nil
}

func (action *BookCreate) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	bcr := &BookCreateRequest{}
	bcr.Title = r.FormValue("title")
	bcr.Author = r.FormValue("author")
	bcr.DateFinished = r.FormValue("dateFinished")
	bcr.Media = r.FormValue("media")

	err := createBook(bcr)
	if err != nil {
		panic(err)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)

}
