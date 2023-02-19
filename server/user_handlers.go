package server

import (
	"net/http"

	"github.com/jackc/booklog/data"
	"github.com/jackc/booklog/view"
)

func UserHome(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(dbconn)
	pathUser := ctx.Value(RequestPathUserKey).(*data.UserMin)

	booksPerYear, err := data.BooksPerYear(ctx, db, pathUser.ID)
	if err != nil {
		InternalServerErrorHandler(w, r, err)
		return
	}

	booksPerMonthForLastYear, err := data.BooksPerMonthForLastYear(ctx, db, pathUser.ID)
	if err != nil {
		InternalServerErrorHandler(w, r, err)
		return
	}

	books, err := data.GetAllBooks(ctx, db, pathUser.ID)
	if err != nil {
		InternalServerErrorHandler(w, r, err)
		return
	}

	yearBooksLists := make([]*view.YearBookList, 0, len(booksPerYear))
	var ybl *view.YearBookList

	for _, book := range books {
		year := book.FinishDate.Year()
		if ybl == nil || year != ybl.Year {
			ybl = &view.YearBookList{Year: year}
			yearBooksLists = append(yearBooksLists, ybl)
		}

		ybl.Books = append(ybl.Books, book)
	}

	err = getRootTmpl().ExecuteTemplate(w, "user_home.html", map[string]any{
		"bva":                      baseViewArgsFromRequest(r),
		"yearBooksLists":           yearBooksLists,
		"booksPerYear":             booksPerYear,
		"booksPerMonthForLastYear": booksPerMonthForLastYear,
	})
	if err != nil {
		InternalServerErrorHandler(w, r, err)
		return
	}
}
