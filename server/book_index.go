package server

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi"
)

func BookIndex(ctx context.Context, e *Endpoint, w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")

	var booksForYears []*BooksForYear
	var booksForYear *BooksForYear
	rows, _ := e.DB.Query(ctx, `select books.id, title, author, finish_date, media
from books
	join users on books.user_id=users.id
where users.username=$1
order by finish_date desc`, username)
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
		e.InternalServerError(w, r, rows.Err())
		return
	}

	err := RenderBookIndex(w, baseViewDataFromRequest(r), booksForYears, username)
	if err != nil {
		e.InternalServerError(w, r, err)
		return
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
