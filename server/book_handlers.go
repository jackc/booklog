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

	"github.com/go-chi/chi/v5"
	"github.com/jackc/booklog/data"
	"github.com/jackc/booklog/myhandler"
	"github.com/jackc/booklog/route"
	"github.com/jackc/booklog/validate"
	"github.com/jackc/booklog/view"
	"github.com/jackc/pgx/v5"
)

type HandlerEnv struct {
}

// TODO -- LazyConn? A wrapper around *pgxpool.Pool that only acquires a *pgx.Conn on demand, but then uses the same one
// for all subsequent calls. Maybe it should not have a direct dependency on *pgxpool.Pool, but instead have functions to acquire and release.

// On second thought, maybe using a database connection is so common that a lazy system isn't the right approach. Maybe
// HandlerEnv should directly acquire and release the conn. If an endpoint doesn't need it a different handler type
// could be used of this type could be configurable.

// But then on third thought, maybe it is better to have a lazy system. Or at least the wrapper concept. The win with
// the wrapper is customizable acquire and release logic such as setting the user for RLS and unsetting it before
// returning it to the pool.

func mountBookHandlers(r chi.Router, config *myhandler.Config[HandlerEnv]) http.Handler {
	r.Method("GET", "/books", myhandler.NewHandler(config, BookIndex))
	r.Method("GET", "/books/new", myhandler.NewHandler(config, BookNew))
	r.Method("POST", "/books", myhandler.NewHandler(config, BookCreate))
	r.Method("GET", "/books/{id}/edit", parseInt64URLParam("id")(myhandler.NewHandler(config, BookEdit)))
	r.Method("GET", "/books/{id}", parseInt64URLParam("id")(myhandler.NewHandler(config, BookShow)))
	r.Method("GET", "/books/{id}/confirm_delete", parseInt64URLParam("id")(myhandler.NewHandler(config, BookConfirmDelete)))
	r.Method("PATCH", "/books/{id}", parseInt64URLParam("id")(myhandler.NewHandler(config, BookUpdate)))
	r.Method("DELETE", "/books/{id}", parseInt64URLParam("id")(myhandler.NewHandler(config, BookDelete)))
	r.Method("GET", "/books/import_csv/form", myhandler.NewHandler(config, BookImportCSVForm))
	r.Method("POST", "/books/import_csv", myhandler.NewHandler(config, BookImportCSV))
	r.Method("GET", "/books.csv", myhandler.NewHandler(config, BookExportCSV))

	return r
}

func BookIndex(ctx context.Context, request *myhandler.Request[HandlerEnv]) error {
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

	return ctx.Value(RequestHTMLTemplateRendererKey).(*view.HTMLTemplateRenderer).ExecuteTemplate(request.ResponseWriter(), "book_index.html", map[string]any{
		"bva":            baseViewArgsFromRequest(request.Request()),
		"yearBooksLists": yearBooksLists,
	})
}

func BookNew(ctx context.Context, request *myhandler.Request[HandlerEnv]) error {
	var form view.BookEditForm
	return ctx.Value(RequestHTMLTemplateRendererKey).(*view.HTMLTemplateRenderer).ExecuteTemplate(request.ResponseWriter(), "book_new.html", map[string]any{
		"bva":  baseViewArgsFromRequest(request.Request()),
		"form": form,
	})
}

func BookCreate(ctx context.Context, request *myhandler.Request[HandlerEnv]) error {
	db := ctx.Value(RequestDBKey).(dbconn)
	pathUser := ctx.Value(RequestPathUserKey).(*data.UserMin)

	form := view.BookEditForm{
		Title:      request.Request().FormValue("title"),
		Author:     request.Request().FormValue("author"),
		FinishDate: request.Request().FormValue("finishDate"),
		Format:     request.Request().FormValue("format"),
		Location:   request.Request().FormValue("location"),
	}
	attrs, verr := form.Parse()
	if verr != nil {
		return ctx.Value(RequestHTMLTemplateRendererKey).(*view.HTMLTemplateRenderer).ExecuteTemplate(request.ResponseWriter(), "book_new.html", map[string]any{
			"bva":  baseViewArgsFromRequest(request.Request()),
			"form": form,
			"verr": verr,
		})
	}
	attrs.UserID = pathUser.ID

	book, err := data.CreateBook(ctx, db, attrs)
	if err != nil {
		var verr validate.Errors
		if errors.As(err, &verr) {
			return ctx.Value(RequestHTMLTemplateRendererKey).(*view.HTMLTemplateRenderer).ExecuteTemplate(request.ResponseWriter(), "book_new.html", map[string]any{
				"bva":  baseViewArgsFromRequest(request.Request()),
				"form": form,
				"verr": verr,
			})
		}
		return err
	}

	http.Redirect(request.ResponseWriter(), request.Request(), route.BookPath(pathUser.Username, book.ID), http.StatusSeeOther)
	return nil
}

func BookConfirmDelete(ctx context.Context, request *myhandler.Request[HandlerEnv]) error {
	db := ctx.Value(RequestDBKey).(dbconn)
	bookID := int64URLParam(request.Request(), "id")

	book, err := data.GetBook(ctx, db, bookID)
	if err != nil {
		var nfErr *data.NotFoundError
		if errors.As(err, &nfErr) {
			NotFoundHandler(request.ResponseWriter(), request.Request())
			return nil
		} else {
			return err
		}
	}

	return ctx.Value(RequestHTMLTemplateRendererKey).(*view.HTMLTemplateRenderer).ExecuteTemplate(request.ResponseWriter(), "book_confirm_delete.html", map[string]any{
		"bva":  baseViewArgsFromRequest(request.Request()),
		"book": book,
	})
}

func BookDelete(ctx context.Context, request *myhandler.Request[HandlerEnv]) error {
	db := ctx.Value(RequestDBKey).(dbconn)
	pathUser := ctx.Value(RequestPathUserKey).(*data.UserMin)
	bookID := int64URLParam(request.Request(), "id")

	err := data.DeleteBook(ctx, db, bookID)
	if err != nil {
		var nfErr *data.NotFoundError
		if errors.As(err, &nfErr) {
			NotFoundHandler(request.ResponseWriter(), request.Request())
			return nil
		} else {
			return err
		}
	}

	http.Redirect(request.ResponseWriter(), request.Request(), route.BooksPath(pathUser.Username), http.StatusSeeOther)
	return nil
}

func BookShow(ctx context.Context, request *myhandler.Request[HandlerEnv]) error {
	db := ctx.Value(RequestDBKey).(dbconn)
	bookID := int64URLParam(request.Request(), "id")

	book, err := data.GetBook(ctx, db, bookID)
	if err != nil {
		var nfErr *data.NotFoundError
		if errors.As(err, &nfErr) {
			NotFoundHandler(request.ResponseWriter(), request.Request())
			return nil
		} else {
			return err
		}
	}

	return ctx.Value(RequestHTMLTemplateRendererKey).(*view.HTMLTemplateRenderer).ExecuteTemplate(request.ResponseWriter(), "book_show.html", map[string]any{
		"bva":  baseViewArgsFromRequest(request.Request()),
		"book": book,
	})
}

func BookEdit(ctx context.Context, request *myhandler.Request[HandlerEnv]) error {
	db := ctx.Value(RequestDBKey).(dbconn)
	bookID := int64URLParam(request.Request(), "id")
	pathUser := ctx.Value(RequestPathUserKey).(*data.UserMin)

	var form view.BookEditForm
	var FinishDate time.Time
	err := db.QueryRow(ctx, "select title, author, finish_date, format, coalesce(location, '') from books where id=$1 and user_id=$2", bookID, pathUser.ID).
		Scan(&form.Title, &form.Author, &FinishDate, &form.Format, &form.Location)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			NotFoundHandler(request.ResponseWriter(), request.Request())
			return nil
		} else {
			return err
		}
	}
	form.FinishDate = FinishDate.Format("2006-01-02")

	return ctx.Value(RequestHTMLTemplateRendererKey).(*view.HTMLTemplateRenderer).ExecuteTemplate(request.ResponseWriter(), "book_edit.html", map[string]any{
		"bva":    baseViewArgsFromRequest(request.Request()),
		"bookID": bookID,
		"form":   form,
	})
}

func BookUpdate(ctx context.Context, request *myhandler.Request[HandlerEnv]) error {
	db := ctx.Value(RequestDBKey).(dbconn)
	bookID := int64URLParam(request.Request(), "id")
	pathUser := ctx.Value(RequestPathUserKey).(*data.UserMin)

	form := view.BookEditForm{
		Title:      request.Request().FormValue("title"),
		Author:     request.Request().FormValue("author"),
		FinishDate: request.Request().FormValue("finishDate"),
		Format:     request.Request().FormValue("format"),
		Location:   request.Request().FormValue("location"),
	}
	attrs, verr := form.Parse()
	if verr != nil {
		return ctx.Value(RequestHTMLTemplateRendererKey).(*view.HTMLTemplateRenderer).ExecuteTemplate(request.ResponseWriter(), "book_edit.html", map[string]any{
			"bva":    baseViewArgsFromRequest(request.Request()),
			"bookID": bookID,
			"form":   form,
			"verr":   verr,
		})
	}
	attrs.ID = bookID

	err := data.UpdateBook(ctx, db, attrs)
	if err != nil {
		var verr validate.Errors
		if errors.As(err, &verr) {
			return ctx.Value(RequestHTMLTemplateRendererKey).(*view.HTMLTemplateRenderer).ExecuteTemplate(request.ResponseWriter(), "book_new.html", map[string]any{
				"bva":    baseViewArgsFromRequest(request.Request()),
				"bookID": bookID,
				"form":   form,
				"verr":   verr,
			})
		}

		var nfErr *data.NotFoundError
		if errors.As(err, &nfErr) {
			NotFoundHandler(request.ResponseWriter(), request.Request())
			return nil
		} else {
			return err
		}
	}

	http.Redirect(request.ResponseWriter(), request.Request(), route.BookPath(pathUser.Username, bookID), http.StatusSeeOther)
	return nil
}

func BookImportCSVForm(ctx context.Context, request *myhandler.Request[HandlerEnv]) error {
	return ctx.Value(RequestHTMLTemplateRendererKey).(*view.HTMLTemplateRenderer).ExecuteTemplate(request.ResponseWriter(), "book_import_csv_form.html", map[string]any{
		"bva": baseViewArgsFromRequest(request.Request()),
	})
}

// TODO - do transactions right

func BookImportCSV(ctx context.Context, request *myhandler.Request[HandlerEnv]) error {
	db := ctx.Value(RequestDBKey).(dbconn)
	pathUser := ctx.Value(RequestPathUserKey).(*data.UserMin)

	request.Request().ParseMultipartForm(10 << 20)

	file, _, err := request.Request().FormFile("file")
	if err != nil {
		return err
	}
	defer file.Close()

	err = importBooksFromCSV(ctx, db, pathUser.ID, file)
	if err != nil {
		return ctx.Value(RequestHTMLTemplateRendererKey).(*view.HTMLTemplateRenderer).ExecuteTemplate(request.ResponseWriter(), "book_import_csv_form.html", map[string]any{
			"bva":       baseViewArgsFromRequest(request.Request()),
			"importErr": err,
		})
	}

	http.Redirect(request.ResponseWriter(), request.Request(), route.BooksPath(pathUser.Username), http.StatusSeeOther)
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

func BookExportCSV(ctx context.Context, request *myhandler.Request[HandlerEnv]) error {
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

	request.ResponseWriter().Header().Set("Content-Type", "text/csv")
	request.ResponseWriter().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=booklog-%s.csv", pathUser.Username))
	_, err := buf.WriteTo(request.ResponseWriter())
	return err
}
