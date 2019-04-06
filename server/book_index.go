package server

import (
	"context"
	"html/template"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/jackc/pgconn"
	"github.com/spf13/viper"
)

type BookIndex struct {
	templates *template.Template
}

func (action *BookIndex) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := pgconn.Connect(context.Background(), viper.GetString("database_uri"))
	if err != nil {
		panic(err)
	}
	defer conn.Close(context.Background())

	result := conn.ExecParams(context.Background(), "select id, title, author, date_finished, media from book order by date_finished asc", nil, nil, nil, nil).Read()
	if result.Err != nil {
		panic(result.Err)
	}

	books := make([]BookRow001, len(result.Rows))
	for i := 0; i < len(books); i++ {
		books[i] = BookRow001{
			ID:           string(result.Rows[i][0]),
			Title:        string(result.Rows[i][1]),
			Author:       string(result.Rows[i][2]),
			DateFinished: string(result.Rows[i][3]),
			Media:        string(result.Rows[i][4]),
		}
	}

	tmpl := action.templates.Lookup("book_index")
	err = tmpl.Execute(w, map[string]interface{}{"Books": books, csrf.TemplateTag: csrf.TemplateField(r)})
	if err != nil {
		panic(err)
	}
}

type BookRow001 struct {
	ID           string
	Title        string
	Author       string
	DateFinished string
	Media        string
}
