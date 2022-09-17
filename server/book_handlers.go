package server

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jackc/booklog/data"
	"github.com/jackc/booklog/route"
	"github.com/jackc/booklog/validate"
	"github.com/jackc/booklog/view"
	"github.com/jackc/pgx/v5"
)

func BookIndex(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(dbconn)
	pathUser := ctx.Value(RequestPathUserKey).(*data.UserMin)

	books, err := data.GetAllBooks(ctx, db, pathUser.ID)
	if err != nil {
		InternalServerErrorHandler(w, r, err)
		return
	}

	yearBooksLists := make([]*view.YearBookList, 0)
	var ybl *view.YearBookList

	for _, book := range books {
		year := book.FinishDate.Year()
		if ybl == nil || year != ybl.Year {
			ybl = &view.YearBookList{Year: year}
			yearBooksLists = append(yearBooksLists, ybl)
		}

		ybl.Books = append(ybl.Books, book)
	}

	err = view.BookIndex(w, baseViewArgsFromRequest(r), yearBooksLists)
	if err != nil {
		InternalServerErrorHandler(w, r, err)
		return
	}
}

func BookNew(w http.ResponseWriter, r *http.Request) {
	var form view.BookEditForm
	err := view.BookNew(w, baseViewArgsFromRequest(r), form, nil)
	if err != nil {
		InternalServerErrorHandler(w, r, err)
		return
	}
}

func BookCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(dbconn)
	pathUser := ctx.Value(RequestPathUserKey).(*data.UserMin)

	form := view.BookEditForm{
		Title:      r.FormValue("title"),
		Author:     r.FormValue("author"),
		FinishDate: r.FormValue("finishDate"),
		Format:     r.FormValue("format"),
		Location:   r.FormValue("location"),
	}
	attrs, verr := form.Parse()
	if verr != nil {
		err := view.BookNew(w, baseViewArgsFromRequest(r), form, verr)
		if err != nil {
			InternalServerErrorHandler(w, r, err)
		}
		return
	}
	attrs.UserID = pathUser.ID

	book, err := data.CreateBook(ctx, db, attrs)
	if err != nil {
		var verr validate.Errors
		if errors.As(err, &verr) {
			err := view.BookNew(w, baseViewArgsFromRequest(r), form, verr)
			if err != nil {
				InternalServerErrorHandler(w, r, err)
			}
			return
		}

		InternalServerErrorHandler(w, r, err)
		return
	}

	http.Redirect(w, r, route.BookPath(pathUser.Username, book.ID), http.StatusSeeOther)
}

func BookConfirmDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(dbconn)
	bookID := int64URLParam(r, "id")

	book, err := data.GetBook(ctx, db, bookID)
	if err != nil {
		var nfErr *data.NotFoundError
		if errors.As(err, &nfErr) {
			NotFoundHandler(w, r)
		} else {
			InternalServerErrorHandler(w, r, err)
		}
		return
	}

	err = view.BookConfirmDelete(w, baseViewArgsFromRequest(r), book)
	if err != nil {
		InternalServerErrorHandler(w, r, err)
		return
	}
}

func BookDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(dbconn)
	pathUser := ctx.Value(RequestPathUserKey).(*data.UserMin)
	bookID := int64URLParam(r, "id")

	err := data.DeleteBook(ctx, db, bookID)
	if err != nil {
		var nfErr *data.NotFoundError
		if errors.As(err, &nfErr) {
			NotFoundHandler(w, r)
		} else {
			InternalServerErrorHandler(w, r, err)
		}
		return
	}

	http.Redirect(w, r, route.BooksPath(pathUser.Username), http.StatusSeeOther)
}

func BookShow(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(dbconn)
	bookID := int64URLParam(r, "id")

	book, err := data.GetBook(ctx, db, bookID)
	if err != nil {
		var nfErr *data.NotFoundError
		if errors.As(err, &nfErr) {
			NotFoundHandler(w, r)
		} else {
			InternalServerErrorHandler(w, r, err)
		}
		return
	}

	err = view.BookShow(w, baseViewArgsFromRequest(r), book)
	if err != nil {
		InternalServerErrorHandler(w, r, err)
		return
	}
}

func BookEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(dbconn)
	pathUser := ctx.Value(RequestPathUserKey).(*data.UserMin)
	bookID := int64URLParam(r, "id")

	var form view.BookEditForm
	var FinishDate time.Time
	err := db.QueryRow(ctx, "select title, author, finish_date, format, coalesce(location, '') from books where id=$1 and user_id=$2", bookID, pathUser.ID).
		Scan(&form.Title, &form.Author, &FinishDate, &form.Format, &form.Location)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			NotFoundHandler(w, r)
		} else {
			InternalServerErrorHandler(w, r, err)
		}
		return
	}
	form.FinishDate = FinishDate.Format("2006-01-02")

	err = view.BookEdit(w, baseViewArgsFromRequest(r), bookID, form, nil)
	if err != nil {
		InternalServerErrorHandler(w, r, err)
		return
	}
}

func BookUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(dbconn)
	pathUser := ctx.Value(RequestPathUserKey).(*data.UserMin)
	bookID := int64URLParam(r, "id")

	form := view.BookEditForm{
		Title:      r.FormValue("title"),
		Author:     r.FormValue("author"),
		FinishDate: r.FormValue("finishDate"),
		Format:     r.FormValue("format"),
		Location:   r.FormValue("location"),
	}
	attrs, verr := form.Parse()
	if verr != nil {
		err := view.BookEdit(w, baseViewArgsFromRequest(r), bookID, form, verr)
		if err != nil {
			InternalServerErrorHandler(w, r, err)
		}
		return
	}
	attrs.ID = bookID

	err := data.UpdateBook(ctx, db, attrs)
	if err != nil {
		var verr validate.Errors
		if errors.As(err, &verr) {
			err := view.BookEdit(w, baseViewArgsFromRequest(r), bookID, form, verr)
			if err != nil {
				InternalServerErrorHandler(w, r, err)
			}
			return
		}

		var nfErr *data.NotFoundError
		if errors.As(err, &nfErr) {
			NotFoundHandler(w, r)
		} else {
			InternalServerErrorHandler(w, r, err)
		}
		return
	}

	http.Redirect(w, r, route.BookPath(pathUser.Username, bookID), http.StatusSeeOther)
}

func BookImportCSVForm(w http.ResponseWriter, r *http.Request) {
	err := view.BookImportCSVForm(w, baseViewArgsFromRequest(r), nil)
	if err != nil {
		InternalServerErrorHandler(w, r, err)
		return
	}
}

// TODO - do transactions right

func BookImportCSV(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	conn := ctx.Value(RequestDBKey).(dbconn)
	pathUser := ctx.Value(RequestPathUserKey).(*data.UserMin)

	r.ParseMultipartForm(10 << 20)

	file, _, err := r.FormFile("file")
	if err != nil {
		InternalServerErrorHandler(w, r, err)
		return
	}
	defer file.Close()

	err = importBooksFromCSV(ctx, conn, pathUser.ID, file)
	if err != nil {
		err := view.BookImportCSVForm(w, baseViewArgsFromRequest(r), err)
		if err != nil {
			InternalServerErrorHandler(w, r, err)
			return
		}
		return
	}

	http.Redirect(w, r, route.BooksPath(pathUser.Username), http.StatusSeeOther)
}

func importBooksFromCSV(ctx context.Context, db dbconn, ownerID int64, r io.Reader) error {
	records, err := csv.NewReader(r).ReadAll()
	if err != nil {
		return err
	}

	if len(records) < 2 {
		return errors.New("CSV must have at least 2 rows")
	}

	if len(records[0]) < 5 {
		return errors.New("CSV must have at least 5 columns")
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for i, record := range records[1:] {
		form := view.BookEditForm{
			Title:      record[0],
			Author:     record[1],
			FinishDate: record[2],
			Format:     record[3],
			Location:   record[4],
		}
		if form.Format == "" {
			form.Format = "text"
		}

		attrs, verr := form.Parse()
		if verr != nil {
			return fmt.Errorf("row %d: %w", i+2, verr)
		}
		attrs.UserID = ownerID

		_, err := data.CreateBook(ctx, tx, attrs)
		if err != nil {
			return fmt.Errorf("row %d: %w", i+2, err)
		}
	}

	return tx.Commit(ctx)
}

func BookExportCSV(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(dbconn)
	pathUser := ctx.Value(RequestPathUserKey).(*data.UserMin)

	buf := &bytes.Buffer{}
	csvWriter := csv.NewWriter(buf)
	csvWriter.Write([]string{"title", "author", "finish_date", "format"})

	rows, _ := db.Query(ctx, `select title, author, finish_date, format
from books
where user_id=$1
order by finish_date desc`, pathUser.ID)
	for rows.Next() {
		var title, author, format string
		var finishDate time.Time
		rows.Scan(&title, &author, &finishDate, &format)
		csvWriter.Write([]string{title, author, finishDate.Format("2006-01-02"), format})
	}
	if rows.Err() != nil {
		InternalServerErrorHandler(w, r, rows.Err())
		return
	}

	csvWriter.Flush()
	if csvWriter.Error() != nil {
		InternalServerErrorHandler(w, r, csvWriter.Error())
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=booklog-%s.csv", pathUser.Username))
	_, err := buf.WriteTo(w)
	if err != nil {
		InternalServerErrorHandler(w, r, err)
		return
	}
}
