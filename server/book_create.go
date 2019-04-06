package server

import (
	"context"
	"html/template"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/jackc/booklog/validate"
	"github.com/jackc/pgconn"
	"github.com/spf13/viper"
)

type BookCreate struct {
	templates *template.Template
}

func createBook(bcr *BookCreateRequest) error {
	v := validate.New()
	v.Presence("title", bcr.Title)
	v.Presence("author", bcr.Author)
	v.Presence("dateFinished", bcr.DateFinished)
	v.Presence("media", bcr.Media)

	if v.Err() != nil {
		return v.Err()
	}

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
		tmpl := action.templates.Lookup("book_new")
		err := tmpl.Execute(w, map[string]interface{}{"fields": bcr, "errors": err, csrf.TemplateTag: csrf.TemplateField(r)})
		if err != nil {
			panic(err)
		}
		return
	}

	http.Redirect(w, r, BooksPath(), http.StatusSeeOther)

}
