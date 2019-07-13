package server

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/gorilla/csrf"
)

type BookIndex struct {
}

func (action *BookIndex) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(queryExecer)
	username := chi.URLParam(r, "username")

	var booksForYears []*BooksForYear
	var booksForYear *BooksForYear
	rows, _ := db.Query(context.Background(), `select finished_book.id, title, author, date_finished, media
from finished_book
	join login_account on finished_book.reader_id=login_account.id
where login_account.username=$1
order by date_finished desc`, username)
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

	err := RenderBookIndex(w, csrf.TemplateField(r), booksForYears, username)
	if err != nil {
		panic(err)
	}
}

type BooksForYear struct {
	Year  int
	Books []BookRow001
}

type BookRow001 struct {
	ID           int64
	Title        string
	Author       string
	DateFinished time.Time
	Media        string
}
