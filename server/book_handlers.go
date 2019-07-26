package server

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jackc/booklog/domain"
	"github.com/jackc/booklog/validate"
	"github.com/jackc/pgx/v4"
	errors "golang.org/x/xerrors"
)

type BookEditForm struct {
	Title        string
	Author       string
	DateFinished string
	Media        string
}

func (f BookEditForm) Parse() (domain.BookAttrs, validate.Errors) {
	var err error
	attrs := domain.BookAttrs{
		Title:  f.Title,
		Author: f.Author,
		Media:  f.Media,
	}
	v := validate.New()

	attrs.DateFinished, err = time.Parse("2006-01-02", f.DateFinished)
	if err != nil {
		attrs.DateFinished, err = time.Parse("1/2/2006", f.DateFinished)
		if err != nil {
			v.Add("dateFinished", errors.New("is not a date"))
		}
	}

	if v.Err() != nil {
		return attrs, v.Err().(validate.Errors)
	}

	return attrs, nil
}

func BookIndex(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(queryExecer)
	pathUser := ctx.Value(RequestPathUserKey).(*minUser)

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

	err := RenderBookIndex(w, baseViewDataFromRequest(r), booksForYears, pathUser.Username)
	if err != nil {
		InternalServerErrorHandler(w, r, err)
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

func BookNew(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pathUser := ctx.Value(RequestPathUserKey).(*minUser)

	var form BookEditForm
	err := RenderBookNew(w, baseViewDataFromRequest(r), form, nil, pathUser.Username)
	if err != nil {
		InternalServerErrorHandler(w, r, err)
		return
	}
}

func BookCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(queryExecer)
	pathUser := ctx.Value(RequestPathUserKey).(*minUser)

	form := BookEditForm{
		Title:        r.FormValue("title"),
		Author:       r.FormValue("author"),
		DateFinished: r.FormValue("dateFinished"),
		Media:        r.FormValue("media"),
	}
	attrs, verr := form.Parse()
	if verr != nil {
		err := RenderBookNew(w, baseViewDataFromRequest(r), form, verr, pathUser.Username)
		if err != nil {
			InternalServerErrorHandler(w, r, err)
		}
		return
	}

	bookID, err := domain.CreateBook(ctx, db, pathUser.ID, attrs)
	if err != nil {
		var verr validate.Errors
		if errors.As(err, &verr) {
			err := RenderBookNew(w, baseViewDataFromRequest(r), form, verr, pathUser.Username)
			if err != nil {
				InternalServerErrorHandler(w, r, err)
			}
			return
		}

		InternalServerErrorHandler(w, r, err)
		return
	}

	http.Redirect(w, r, BookPath(pathUser.Username, bookID), http.StatusSeeOther)
}

func BookConfirmDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(queryExecer)
	pathUser := ctx.Value(RequestPathUserKey).(*minUser)
	bookID := int64URLParam(r, "id")

	book, err := domain.GetBook(ctx, db, bookID)
	if err != nil {
		var nfErr domain.NotFoundError
		if errors.As(err, nfErr) {
			NotFoundHandler(w, r)
		} else {
			InternalServerErrorHandler(w, r, err)
		}
		return
	}

	err = RenderBookConfirmDelete(w, baseViewDataFromRequest(r), book, pathUser.Username)
	if err != nil {
		InternalServerErrorHandler(w, r, err)
		return
	}
}

func BookDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(queryExecer)
	pathUser := ctx.Value(RequestPathUserKey).(*minUser)
	bookID := int64URLParam(r, "id")

	err := domain.DeleteBook(ctx, db, bookID)
	if err != nil {
		var nfErr domain.NotFoundError
		if errors.As(err, nfErr) {
			NotFoundHandler(w, r)
		} else {
			InternalServerErrorHandler(w, r, err)
		}
		return
	}

	http.Redirect(w, r, BooksPath(pathUser.Username), http.StatusSeeOther)
}

func BookShow(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(queryExecer)
	pathUser := ctx.Value(RequestPathUserKey).(*minUser)
	bookID := int64URLParam(r, "id")

	book, err := domain.GetBook(ctx, db, bookID)
	if err != nil {
		var nfErr domain.NotFoundError
		if errors.As(err, nfErr) {
			NotFoundHandler(w, r)
		} else {
			InternalServerErrorHandler(w, r, err)
		}
		return
	}

	err = RenderBookShow(w, baseViewDataFromRequest(r), book, pathUser.Username)
	if err != nil {
		InternalServerErrorHandler(w, r, err)
		return
	}
}

func BookEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(queryExecer)
	pathUser := ctx.Value(RequestPathUserKey).(*minUser)
	bookID := int64URLParam(r, "id")

	var form BookEditForm
	var dateFinished time.Time
	err := db.QueryRow(ctx, "select title, author, finish_date, media from books where id=$1 and user_id=$2", bookID, pathUser.ID).
		Scan(&form.Title, &form.Author, &dateFinished, &form.Media)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			NotFoundHandler(w, r)
		} else {
			InternalServerErrorHandler(w, r, err)
		}
		return
	}
	form.DateFinished = dateFinished.Format("2006-01-02")

	err = RenderBookEdit(w, baseViewDataFromRequest(r), bookID, form, nil, pathUser.Username)
	if err != nil {
		InternalServerErrorHandler(w, r, err)
		return
	}
}

func BookUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(queryExecer)
	pathUser := ctx.Value(RequestPathUserKey).(*minUser)
	bookID := int64URLParam(r, "id")

	form := BookEditForm{
		Title:        r.FormValue("title"),
		Author:       r.FormValue("author"),
		DateFinished: r.FormValue("dateFinished"),
		Media:        r.FormValue("media"),
	}
	attrs, verr := form.Parse()
	if verr != nil {
		err := RenderBookEdit(w, baseViewDataFromRequest(r), bookID, form, verr, pathUser.Username)
		if err != nil {
			InternalServerErrorHandler(w, r, err)
		}
		return
	}

	err := domain.UpdateBook(ctx, db, bookID, attrs)
	if err != nil {
		var verr validate.Errors
		if errors.As(err, &verr) {
			err := RenderBookEdit(w, baseViewDataFromRequest(r), bookID, form, verr, pathUser.Username)
			if err != nil {
				InternalServerErrorHandler(w, r, err)
			}
			return
		}

		var nfErr domain.NotFoundError
		if errors.As(err, nfErr) {
			NotFoundHandler(w, r)
		} else {
			InternalServerErrorHandler(w, r, err)
		}
		return
	}

	http.Redirect(w, r, BookPath(pathUser.Username, bookID), http.StatusSeeOther)
}

func BookImportCSVForm(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pathUser := ctx.Value(RequestPathUserKey).(*minUser)

	err := RenderBookImportCSVForm(w, baseViewDataFromRequest(r), pathUser.Username)
	if err != nil {
		InternalServerErrorHandler(w, r, err)
		return
	}
}

func BookImportCSV(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(queryExecer)
	pathUser := ctx.Value(RequestPathUserKey).(*minUser)

	r.ParseMultipartForm(10 << 20)

	file, _, err := r.FormFile("file")
	if err != nil {
		InternalServerErrorHandler(w, r, err)
		return
	}
	defer file.Close()

	err = importBooksFromCSV(ctx, db, pathUser.ID, file)
	if err != nil {
		InternalServerErrorHandler(w, r, err)
		return
	}

	http.Redirect(w, r, BooksPath(pathUser.Username), http.StatusSeeOther)
}

// TODO - need DB transaction control - so queryExecer is insufficient
func importBooksFromCSV(ctx context.Context, db queryExecer, ownerID int64, r io.Reader) error {
	records, err := csv.NewReader(r).ReadAll()
	if err != nil {
		return err
	}

	if len(records) < 2 {
		return errors.New("CSV must have at least 2 rows")
	}

	if len(records[0]) < 4 {
		return errors.New("CSV must have at least 4 columns")
	}

	for i, record := range records[1:] {
		form := BookEditForm{
			Title:        record[0],
			Author:       record[1],
			DateFinished: record[2],
			Media:        record[3],
		}
		if form.Media == "" {
			form.Media = "book"
		}

		attrs, verr := form.Parse()
		if verr != nil {
			return errors.Errorf("row %d: %w", i+1, verr)
		}

		_, err := domain.CreateBook(ctx, db, ownerID, attrs)
		if err != nil {
			return errors.Errorf("row %d: %w", i+1, err)
		}
	}

	return nil
}

func BookExportCSV(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(queryExecer)
	pathUser := ctx.Value(RequestPathUserKey).(*minUser)

	buf := &bytes.Buffer{}
	csvWriter := csv.NewWriter(buf)
	csvWriter.Write([]string{"title", "author", "finish_date", "media"})

	rows, _ := db.Query(ctx, `select title, author, finish_date, media
from books
where user_id=$1
order by finish_date desc`, pathUser.ID)
	for rows.Next() {
		var title, author, media string
		var finishDate time.Time
		rows.Scan(&title, &author, &finishDate, &media)
		csvWriter.Write([]string{title, author, finishDate.Format("2006-01-02"), media})
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
