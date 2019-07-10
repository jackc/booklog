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

	var booksForYears []*BooksForYear
	var booksForYear *BooksForYear
	rows, _ := db.Query(context.Background(), "select id, title, author, date_finished, media from book order by date_finished desc")
	for rows.Next() {
		var b BookRow001
		rows.Scan(&b.ID, &b.Title, &b.Author, &b.DateFinished, &b.Media)
		year := b.DateFinished.Year()
		if booksForYear == nil || year != booksForYear.Year {
			booksForYear = &BooksForYear{Year: year}
			booksForYears = append(booksForYears, booksForYear)
		}

		booksForYear.Books = append(booksForYear.Books, b)
	}
	if rows.Err() != nil {
		panic(rows.Err())
	}

	err := RenderBookIndex(w, csrf.TemplateField(r), booksForYears)
	if err != nil {
		panic(err)
	}
}

type BooksForYear struct {
	Year  int
	Books []BookRow001
}

type BookRow001 struct {
	ID           string
	Title        string
	Author       string
	DateFinished time.Time
	Media        string
}
