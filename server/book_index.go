package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/csrf"
)

type BookIndex struct {
}

func (action *BookIndex) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(queryExecer)

	var books []BookRow001
	rows, _ := db.Query(context.Background(), "select id, title, author, date_finished, media from book order by date_finished asc")
	for rows.Next() {
		var b BookRow001
		rows.Scan(&b.ID, &b.Title, &b.Author, &b.DateFinished, &b.Media)
		books = append(books, b)
	}
	if rows.Err() != nil {
		panic(rows.Err())
	}

	err := RenderBookIndex(w, csrf.TemplateField(r), books)
	if err != nil {
		panic(err)
	}
}

type BookRow001 struct {
	ID           string
	Title        string
	Author       string
	DateFinished time.Time
	Media        string
}
