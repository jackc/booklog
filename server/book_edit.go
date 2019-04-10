package server

import (
	"context"
	"html/template"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/gorilla/csrf"
	"github.com/jackc/pgx"
	"github.com/spf13/viper"
)

type BookEdit struct {
	templates *template.Template
}

func (action *BookEdit) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := pgx.Connect(context.Background(), viper.GetString("database_uri"))
	if err != nil {
		panic(err)
	}
	defer conn.Close(context.Background())

	bcr := &BookCreateRequest{}
	err = conn.QueryRow(context.Background(), "select title, author, date_finished::text, media from book where id=$1", chi.URLParam(r, "id")).Scan(&bcr.Title, &bcr.Author, &bcr.DateFinished, &bcr.Media)
	// TODO - handle not found error
	// if len(result.Rows) == 0 {
	// 	http.NotFound(w, r)
	// 	return
	// }
	if err != nil {
		panic(err)
	}

	tmpl := action.templates.Lookup("book_edit")
	err = tmpl.Execute(w, map[string]interface{}{"bookID": chi.URLParam(r, "id"), "fields": bcr, "errors": map[string]string{}, csrf.TemplateTag: csrf.TemplateField(r)})
	if err != nil {
		panic(err)
	}

}
