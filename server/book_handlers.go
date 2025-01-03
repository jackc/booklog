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
	"github.com/jackc/booklog/view"
	"github.com/jackc/errortree"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/structify"
)

// TODO -- LazyConn? A wrapper around *pgxpool.Pool that only acquires a *pgx.Conn on demand, but then uses the same one
// for all subsequent calls. Maybe it should not have a direct dependency on *pgxpool.Pool, but instead have functions to acquire and release.

// On second thought, maybe using a database connection is so common that a lazy system isn't the right approach. Maybe
// HandlerEnv should directly acquire and release the conn. If an endpoint doesn't need it a different handler type
// could be used of this type could be configurable.

// But then on third thought, maybe it is better to have a lazy system. Or at least the wrapper concept. The win with
// the wrapper is customizable acquire and release logic such as setting the user for RLS and unsetting it before
// returning it to the pool.

func BookIndex(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]any) error {
	db := ctx.Value(RequestDBKey).(dbconn)
	pathUser := ctx.Value(RequestPathUserKey).(*data.UserMin)

	books, err := data.GetAllBooks(ctx, db, pathUser.ID)
	if err != nil {
		return err
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

	return ctx.Value(RequestHTMLTemplateRendererKey).(*view.HTMLTemplateRenderer).ExecuteTemplate(w, "book_index.html", map[string]any{
		"bva":            baseViewArgsFromRequest(r),
		"yearBooksLists": yearBooksLists,
	})
}

func BookNew(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]any) error {
	var form view.BookEditForm
	return ctx.Value(RequestHTMLTemplateRendererKey).(*view.HTMLTemplateRenderer).ExecuteTemplate(w, "book_new.html", map[string]any{
		"bva":  baseViewArgsFromRequest(r),
		"form": form,
	})
}

func BookCreate(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]any) error {
	db := ctx.Value(RequestDBKey).(dbconn)
	pathUser := ctx.Value(RequestPathUserKey).(*data.UserMin)

	var form view.BookEditForm
	_ = structify.Parse(params, &form)
	attrs, verr := form.Parse()
	if verr != nil {
		return ctx.Value(RequestHTMLTemplateRendererKey).(*view.HTMLTemplateRenderer).ExecuteTemplate(w, "book_new.html", map[string]any{
			"bva":  baseViewArgsFromRequest(r),
			"form": form,
			"verr": verr,
		})
	}
	attrs.UserID = pathUser.ID

	book, err := data.CreateBook(ctx, db, attrs)
	if err != nil {
		var verr *errortree.Node
		if errors.As(err, &verr) {
			return ctx.Value(RequestHTMLTemplateRendererKey).(*view.HTMLTemplateRenderer).ExecuteTemplate(w, "book_new.html", map[string]any{
				"bva":  baseViewArgsFromRequest(r),
				"form": form,
				"verr": verr,
			})
		}
		return err
	}

	http.Redirect(w, r, route.BookPath(pathUser.Username, book.ID), http.StatusSeeOther)
	return nil
}

func BookConfirmDelete(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]any) error {
	db := ctx.Value(RequestDBKey).(dbconn)
	bookID := int64URLParam(r, "id")

	book, err := data.GetBook(ctx, db, bookID)
	if err != nil {
		var nfErr *data.NotFoundError
		if errors.As(err, &nfErr) {
			NotFoundHandler(w, r)
			return nil
		} else {
			return err
		}
	}

	return ctx.Value(RequestHTMLTemplateRendererKey).(*view.HTMLTemplateRenderer).ExecuteTemplate(w, "book_confirm_delete.html", map[string]any{
		"bva":  baseViewArgsFromRequest(r),
		"book": book,
	})
}

func BookDelete(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]any) error {
	db := ctx.Value(RequestDBKey).(dbconn)
	pathUser := ctx.Value(RequestPathUserKey).(*data.UserMin)
	bookID := int64URLParam(r, "id")

	err := data.DeleteBook(ctx, db, bookID)
	if err != nil {
		var nfErr *data.NotFoundError
		if errors.As(err, &nfErr) {
			NotFoundHandler(w, r)
			return nil
		} else {
			return err
		}
	}

	http.Redirect(w, r, route.BooksPath(pathUser.Username), http.StatusSeeOther)
	return nil
}

func BookShow(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]any) error {
	db := ctx.Value(RequestDBKey).(dbconn)
	bookID := int64URLParam(r, "id")

	book, err := data.GetBook(ctx, db, bookID)
	if err != nil {
		var nfErr *data.NotFoundError
		if errors.As(err, &nfErr) {
			NotFoundHandler(w, r)
			return nil
		} else {
			return err
		}
	}

	return ctx.Value(RequestHTMLTemplateRendererKey).(*view.HTMLTemplateRenderer).ExecuteTemplate(w, "book_show.html", map[string]any{
		"bva":  baseViewArgsFromRequest(r),
		"book": book,
	})
}

func BookEdit(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]any) error {
	db := ctx.Value(RequestDBKey).(dbconn)
	bookID := int64URLParam(r, "id")
	pathUser := ctx.Value(RequestPathUserKey).(*data.UserMin)

	var form view.BookEditForm
	var FinishDate time.Time
	err := db.QueryRow(ctx, "select title, author, finish_date, format, coalesce(location, '') from books where id=$1 and user_id=$2", bookID, pathUser.ID).
		Scan(&form.Title, &form.Author, &FinishDate, &form.Format, &form.Location)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			NotFoundHandler(w, r)
			return nil
		} else {
			return err
		}
	}
	form.FinishDate = FinishDate.Format("2006-01-02")

	return ctx.Value(RequestHTMLTemplateRendererKey).(*view.HTMLTemplateRenderer).ExecuteTemplate(w, "book_edit.html", map[string]any{
		"bva":    baseViewArgsFromRequest(r),
		"bookID": bookID,
		"form":   form,
	})
}

func BookUpdate(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]any) error {
	db := ctx.Value(RequestDBKey).(dbconn)
	bookID := int64URLParam(r, "id")
	pathUser := ctx.Value(RequestPathUserKey).(*data.UserMin)

	var form view.BookEditForm
	_ = structify.Parse(params, &form)
	attrs, verr := form.Parse()
	if verr != nil {
		return ctx.Value(RequestHTMLTemplateRendererKey).(*view.HTMLTemplateRenderer).ExecuteTemplate(w, "book_edit.html", map[string]any{
			"bva":    baseViewArgsFromRequest(r),
			"bookID": bookID,
			"form":   form,
			"verr":   verr,
		})
	}
	attrs.ID = bookID

	err := data.UpdateBook(ctx, db, attrs)
	if err != nil {
		var verr *errortree.Node
		if errors.As(err, &verr) {
			return ctx.Value(RequestHTMLTemplateRendererKey).(*view.HTMLTemplateRenderer).ExecuteTemplate(w, "book_new.html", map[string]any{
				"bva":    baseViewArgsFromRequest(r),
				"bookID": bookID,
				"form":   form,
				"verr":   verr,
			})
		}

		var nfErr *data.NotFoundError
		if errors.As(err, &nfErr) {
			NotFoundHandler(w, r)
			return nil
		} else {
			return err
		}
	}

	http.Redirect(w, r, route.BookPath(pathUser.Username, bookID), http.StatusSeeOther)
	return nil
}

func BookImportCSVForm(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]any) error {
	return ctx.Value(RequestHTMLTemplateRendererKey).(*view.HTMLTemplateRenderer).ExecuteTemplate(w, "book_import_csv_form.html", map[string]any{
		"bva": baseViewArgsFromRequest(r),
	})
}

// TODO - do transactions right

func BookImportCSV(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]any) error {
	db := ctx.Value(RequestDBKey).(dbconn)
	pathUser := ctx.Value(RequestPathUserKey).(*data.UserMin)

	r.ParseMultipartForm(10 << 20)

	file, _, err := r.FormFile("file")
	if err != nil {
		return err
	}
	defer file.Close()

	err = importBooksFromCSV(ctx, db, pathUser.ID, file)
	if err != nil {
		return ctx.Value(RequestHTMLTemplateRendererKey).(*view.HTMLTemplateRenderer).ExecuteTemplate(w, "book_import_csv_form.html", map[string]any{
			"bva":       baseViewArgsFromRequest(r),
			"importErr": err,
		})
	}

	http.Redirect(w, r, route.BooksPath(pathUser.Username), http.StatusSeeOther)
	return nil
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

func BookExportCSV(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]any) error {
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
		return rows.Err()
	}

	csvWriter.Flush()
	if csvWriter.Error() != nil {
		return csvWriter.Error()
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=booklog-%s.csv", pathUser.Username))
	_, err := buf.WriteTo(w)
	return err
}
