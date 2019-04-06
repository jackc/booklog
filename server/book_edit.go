package server

import (
	"context"
	"html/template"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/gorilla/csrf"
	"github.com/jackc/pgconn"
	"github.com/spf13/viper"
)

type BookEdit struct {
	templates *template.Template
}

func (action *BookEdit) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := pgconn.Connect(context.Background(), viper.GetString("database_uri"))
	if err != nil {
		panic(err)
	}
	defer conn.Close(context.Background())

	result := conn.ExecParams(context.Background(), "select title, author, date_finished, media from book where id=$1", [][]byte{[]byte(chi.URLParam(r, "id"))}, nil, nil, nil).Read()
	if result.Err != nil {
		panic(result.Err)
	}

	if len(result.Rows) == 0 {
		http.NotFound(w, r)
		return
	}

	bcr := &BookCreateRequest{
		Title:        string(result.Rows[0][0]),
		Author:       string(result.Rows[0][1]),
		DateFinished: string(result.Rows[0][2]),
		Media:        string(result.Rows[0][3]),
	}

	tmpl := action.templates.Lookup("book_edit")
	err = tmpl.Execute(w, map[string]interface{}{"bookID": chi.URLParam(r, "id"), "fields": bcr, "errors": map[string]string{}, csrf.TemplateTag: csrf.TemplateField(r)})
	if err != nil {
		panic(err)
	}

}
