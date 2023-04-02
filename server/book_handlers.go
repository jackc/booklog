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
	"github.com/jackc/structify"
)

type HandlerEnv struct {
	request *myhandler.Request[HandlerEnv]

	dbconn *pgx.Conn
}

// TODO -- LazyConn? A wrapper around *pgxpool.Pool that only acquires a *pgx.Conn on demand, but then uses the same one
// for all subsequent calls. Maybe it should not have a direct dependency on *pgxpool.Pool, but instead have functions to acquire and release.

func (env *HandlerEnv) DBConn() *pgx.Conn {
	if env.dbconn == nil {
		// TODO
	}
	return env.dbconn
}

func mountBookHandlers(r chi.Router, appServer *AppServer) http.Handler {
	config := &myhandler.Config[HandlerEnv]{
		HTMLTemplateRenderer: appServer.htr,

		BuildEnv: func(ctx context.Context, request *myhandler.Request[HandlerEnv]) (*HandlerEnv, error) {
			return &HandlerEnv{
				request: request,
			}, nil
		},
		CleanupEnv: func(ctx context.Context, request *myhandler.Request[HandlerEnv]) error {
			return nil
		},
	}

	r.Method("GET", "/books", myhandler.NewHandler(config, BookIndex))
	r.Method("GET", "/books/new", http.HandlerFunc(BookNew))
	r.Method("POST", "/books", http.HandlerFunc(BookCreate))
	r.Method("GET", "/books/{id}/edit", parseInt64URLParam("id")(http.HandlerFunc(BookEdit)))
	r.Method("GET", "/books/{id}", parseInt64URLParam("id")(http.HandlerFunc(BookShow)))
	r.Method("GET", "/books/{id}/confirm_delete", parseInt64URLParam("id")(http.HandlerFunc(BookConfirmDelete)))
	r.Method("PATCH", "/books/{id}", parseInt64URLParam("id")(http.HandlerFunc(BookUpdate)))
	r.Method("DELETE", "/books/{id}", parseInt64URLParam("id")(http.HandlerFunc(BookDelete)))
	r.Method("GET", "/books/import_csv/form", http.HandlerFunc(BookImportCSVForm))
	r.Method("POST", "/books/import_csv", http.HandlerFunc(BookImportCSV))
	r.Method("GET", "/books.csv", http.HandlerFunc(BookExportCSV))

	return r
}

func BookIndex(ctx context.Context, request *myhandler.Request[HandlerEnv]) error {
	// db := request.Env().DB()
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

	return request.RenderHTMLTemplate("book_index.html", map[string]any{
		"bva":            baseViewArgsFromRequest(request.Request()),
		"yearBooksLists": yearBooksLists,
	})
}

func BookNew(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	htr := ctx.Value(RequestHTMLTemplateRendererKey).(*view.HTMLTemplateRenderer)
	var form view.BookEditForm
	err := htr.ExecuteTemplate(w, "book_new.html", map[string]any{
		"bva":  baseViewArgsFromRequest(r),
		"form": form,
	})
	if err != nil {
		InternalServerErrorHandler(w, r, err)
		return
	}
}

func BookCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(dbconn)
	htr := ctx.Value(RequestHTMLTemplateRendererKey).(*view.HTMLTemplateRenderer)
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
		err := htr.ExecuteTemplate(w, "book_new.html", map[string]any{
			"bva":  baseViewArgsFromRequest(r),
			"form": form,
			"verr": verr,
		})
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
			err := htr.ExecuteTemplate(w, "book_new.html", map[string]any{
				"bva":  baseViewArgsFromRequest(r),
				"form": form,
				"verr": verr,
			})
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
	htr := ctx.Value(RequestHTMLTemplateRendererKey).(*view.HTMLTemplateRenderer)
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

	err = htr.ExecuteTemplate(w, "book_confirm_delete.html", map[string]any{
		"bva":  baseViewArgsFromRequest(r),
		"book": book,
	})
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
	htr := ctx.Value(RequestHTMLTemplateRendererKey).(*view.HTMLTemplateRenderer)
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

	err = htr.ExecuteTemplate(w, "book_show.html", map[string]any{
		"bva":  baseViewArgsFromRequest(r),
		"book": book,
	})
	if err != nil {
		InternalServerErrorHandler(w, r, err)
		return
	}
}

func BookEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(dbconn)
	htr := ctx.Value(RequestHTMLTemplateRendererKey).(*view.HTMLTemplateRenderer)
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

	err = htr.ExecuteTemplate(w, "book_edit.html", map[string]any{
		"bva":    baseViewArgsFromRequest(r),
		"bookID": bookID,
		"form":   form,
	})
	if err != nil {
		InternalServerErrorHandler(w, r, err)
		return
	}
}

func BookUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(dbconn)
	htr := ctx.Value(RequestHTMLTemplateRendererKey).(*view.HTMLTemplateRenderer)
	pathUser := ctx.Value(RequestPathUserKey).(*data.UserMin)
	bookID := int64URLParam(r, "id")

	params := ctx.Value(RequestParamsKey).(map[string]any)
	var form view.BookEditForm
	err := structify.Parse(params, &form)
	if err != nil {
		InternalServerErrorHandler(w, r, err)
		return
	}
	attrs, verr := form.Parse()
	if verr != nil {
		err := htr.ExecuteTemplate(w, "book_edit.html", map[string]any{
			"bva":    baseViewArgsFromRequest(r),
			"bookID": bookID,
			"form":   form,
			"verr":   verr,
		})
		if err != nil {
			InternalServerErrorHandler(w, r, err)
		}
		return
	}
	attrs.ID = bookID

	err = data.UpdateBook(ctx, db, attrs)
	if err != nil {
		var verr validate.Errors
		if errors.As(err, &verr) {
			err := htr.ExecuteTemplate(w, "book_new.html", map[string]any{
				"bva":    baseViewArgsFromRequest(r),
				"bookID": bookID,
				"form":   form,
				"verr":   verr,
			})
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
	ctx := r.Context()
	htr := ctx.Value(RequestHTMLTemplateRendererKey).(*view.HTMLTemplateRenderer)
	err := htr.ExecuteTemplate(w, "book_import_csv_form.html", map[string]any{
		"bva": baseViewArgsFromRequest(r),
	})
	if err != nil {
		InternalServerErrorHandler(w, r, err)
		return
	}
}

// TODO - do transactions right

func BookImportCSV(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	conn := ctx.Value(RequestDBKey).(dbconn)
	htr := ctx.Value(RequestHTMLTemplateRendererKey).(*view.HTMLTemplateRenderer)
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
		err := htr.ExecuteTemplate(w, "book_import_csv_form.html", map[string]any{
			"bva":       baseViewArgsFromRequest(r),
			"importErr": err,
		})
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
