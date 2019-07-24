package server

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/jackc/booklog/domain"
	"github.com/jackc/booklog/validate"
	"github.com/jackc/pgx/v4"
	errors "golang.org/x/xerrors"
)

func BookIndex(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(queryExecer)
	username := chi.URLParam(r, "username")

	var booksForYears []*BooksForYear
	var booksForYear *BooksForYear
	rows, _ := db.Query(ctx, `select books.id, title, author, finish_date, media
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
		InternalServerErrorHandler(w, r, rows.Err())
		return
	}

	err := RenderBookIndex(w, baseViewDataFromRequest(r), booksForYears, username)
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
	username := chi.URLParam(r, "username")

	var createBookArgs domain.CreateBookArgs
	err := RenderBookNew(w, baseViewDataFromRequest(r), createBookArgs, nil, username)
	if err != nil {
		InternalServerErrorHandler(w, r, err)
		return
	}
}

func BookCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(queryExecer)

	username := chi.URLParam(r, "username")
	var userID int64
	err := db.QueryRow(ctx, "select id from users where username=$1", username).Scan(&userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			NotFoundHandler(w, r)
		} else {
			InternalServerErrorHandler(w, r, err)
		}
		return
	}

	cba := domain.CreateBookArgs{
		ReaderID:     userID,
		Title:        r.FormValue("title"),
		Author:       r.FormValue("author"),
		DateFinished: r.FormValue("dateFinished"),
		Media:        r.FormValue("media"),
	}

	err = domain.CreateBook(ctx, db, cba)
	if err != nil {
		var verr validate.Errors
		if errors.As(err, &verr) {
			err := RenderBookNew(w, baseViewDataFromRequest(r), cba, verr, username)
			if err != nil {
				InternalServerErrorHandler(w, r, err)
			}
			return
		}

		InternalServerErrorHandler(w, r, err)
		return
	}

	http.Redirect(w, r, BooksPath(username), http.StatusSeeOther)
}

func BookDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(queryExecer)

	bookID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		NotFoundHandler(w, r)
		return
	}

	var username string
	err = db.QueryRow(ctx, "select username from users join books on users.id=books.user_id where books.id=$1", bookID).Scan(&username)
	if err != nil {
		NotFoundHandler(w, r)
		return
	}

	dba := domain.DeleteBookArgs{
		ID: bookID,
	}

	err = domain.DeleteBook(ctx, db, dba)
	if err != nil {
		InternalServerErrorHandler(w, r, err)
		return
	}

	http.Redirect(w, r, BooksPath(username), http.StatusSeeOther)
}

func BookEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(queryExecer)

	username := chi.URLParam(r, "username")
	bookID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		NotFoundHandler(w, r)
		return
	}

	uba := domain.UpdateBookArgs{}
	err = db.QueryRow(ctx, "select title, author, finish_date::text, media from books where id=$1", bookID).Scan(&uba.Title, &uba.Author, &uba.DateFinished, &uba.Media)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			NotFoundHandler(w, r)
		} else {
			InternalServerErrorHandler(w, r, err)
		}
		return
	}

	err = RenderBookEdit(w, baseViewDataFromRequest(r), bookID, uba, nil, username)
	if err != nil {
		InternalServerErrorHandler(w, r, err)
		return
	}
}

func BookUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(queryExecer)

	bookID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		NotFoundHandler(w, r)
		return
	}

	var username string
	err = db.QueryRow(ctx, "select username from users join books on users.id=books.user_id where books.id=$1", bookID).Scan(&username)
	if err != nil {
		NotFoundHandler(w, r)
		return
	}

	uba := domain.UpdateBookArgs{
		ID:           bookID,
		Title:        r.FormValue("title"),
		Author:       r.FormValue("author"),
		DateFinished: r.FormValue("dateFinished"),
		Media:        r.FormValue("media"),
	}

	err = domain.UpdateBook(ctx, db, uba)
	if err != nil {
		var verr validate.Errors
		if errors.As(err, &verr) {
			err := RenderBookEdit(w, baseViewDataFromRequest(r), bookID, uba, verr, username)
			if err != nil {
				InternalServerErrorHandler(w, r, err)
			}
			return
		}

		InternalServerErrorHandler(w, r, err)
		return
	}

	http.Redirect(w, r, BooksPath(username), http.StatusSeeOther)
}

func BookImportCSVForm(w http.ResponseWriter, r *http.Request) {
	err := RenderBookImportCSVForm(w, baseViewDataFromRequest(r), chi.URLParam(r, "username"))
	if err != nil {
		InternalServerErrorHandler(w, r, err)
		return
	}
}

func BookImportCSV(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(queryExecer)

	username := chi.URLParam(r, "username")
	var userID int64
	err := db.QueryRow(ctx, "select id from users where username=$1", username).Scan(&userID)
	if err != nil {
		NotFoundHandler(w, r)
		return
	}

	r.ParseMultipartForm(10 << 20)

	file, _, err := r.FormFile("file")
	if err != nil {
		InternalServerErrorHandler(w, r, err)
		return
	}
	defer file.Close()

	err = domain.ImportBooksFromCSV(ctx, db, userID, file)
	if err != nil {
		InternalServerErrorHandler(w, r, err)
		return
	}

	http.Redirect(w, r, BooksPath(username), http.StatusSeeOther)
}
