package server

import (
	"net/http"

	"github.com/jackc/booklog/data"
)

func UserHome(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(queryExecer)
	pathUser := ctx.Value(RequestPathUserKey).(*minUser)

	booksPerYear, err := data.BooksPerYear(ctx, db, pathUser.ID)

	var booksForYears []*BooksForYear
	var booksForYear *BooksForYear
	rows, _ := db.Query(ctx, `select books.id, title, author, finish_date, media
from books
where user_id=$1
order by finish_date desc`, pathUser.ID)
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
		InternalServerErrorHandler(w, r, rows.Err())
		return
	}

	err = RenderUserHome(w, baseViewDataFromRequest(r), booksForYears, booksPerYear, pathUser.Username)
	if err != nil {
		InternalServerErrorHandler(w, r, err)
		return
	}
}
