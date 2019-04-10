package server

import (
	"context"
	"errors"
	"html/template"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/gorilla/csrf"
	"github.com/jackc/booklog/validate"
	"github.com/jackc/pgx"
	"github.com/spf13/viper"
)

type BookUpdate struct {
	templates *template.Template
}

func updateBook(id string, bcr *BookCreateRequest) error {
	v := validate.New()
	v.Presence("title", bcr.Title)
	v.Presence("author", bcr.Author)
	v.Presence("dateFinished", bcr.DateFinished)
	v.Presence("media", bcr.Media)

	if v.Err() != nil {
		return v.Err()
	}

	conn, err := pgx.Connect(context.Background(), viper.GetString("database_uri"))
	if err != nil {
		panic(err)
	}
	defer conn.Close(context.Background())

	commandTag, err := conn.Exec(context.Background(), "update book set title=$1, author=$2, date_finished=$3, media=$4 where id=$5", bcr.Title, bcr.Author, bcr.DateFinished, bcr.Media, id)
	if err != nil {
		return err
	}
	if commandTag != "UPDATE 1" {
		return errors.New("not found")
	}

	return nil
}

func (action *BookUpdate) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	bcr := &BookCreateRequest{}
	bcr.Title = r.FormValue("title")
	bcr.Author = r.FormValue("author")
	bcr.DateFinished = r.FormValue("dateFinished")
	bcr.Media = r.FormValue("media")

	err := updateBook(chi.URLParam(r, "id"), bcr)
	if err != nil {
		tmpl := action.templates.Lookup("book_edit")
		// TODO - if errors is not a map this fails
		err := tmpl.Execute(w, map[string]interface{}{"bookID": chi.URLParam(r, "id"), "fields": bcr, "errors": err, csrf.TemplateTag: csrf.TemplateField(r)})
		if err != nil {
			panic(err)
		}
		return
	}

	http.Redirect(w, r, BooksPath(), http.StatusSeeOther)

}
