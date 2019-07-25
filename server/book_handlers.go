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

	var createBookArgs domain.CreateBookArgs
	err := RenderBookNew(w, baseViewDataFromRequest(r), createBookArgs, nil, pathUser.Username)
	if err != nil {
		InternalServerErrorHandler(w, r, err)
		return
	}
}

func BookCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(queryExecer)
	pathUser := ctx.Value(RequestPathUserKey).(*minUser)

	cba := domain.CreateBookArgs{
		ReaderID:     pathUser.ID,
		Title:        r.FormValue("title"),
		Author:       r.FormValue("author"),
		DateFinished: r.FormValue("dateFinished"),
		Media:        r.FormValue("media"),
	}

	err := domain.CreateBook(ctx, db, cba)
	if err != nil {
		var verr validate.Errors
		if errors.As(err, &verr) {
			err := RenderBookNew(w, baseViewDataFromRequest(r), cba, verr, pathUser.Username)
			if err != nil {
				InternalServerErrorHandler(w, r, err)
			}
			return
		}

		InternalServerErrorHandler(w, r, err)
		return
	}

	http.Redirect(w, r, BooksPath(pathUser.Username), http.StatusSeeOther)
}

func BookDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(queryExecer)
	session := ctx.Value(RequestSessionKey).(*Session)
	pathUser := ctx.Value(RequestPathUserKey).(*minUser)

	err := domain.DeleteBookParse(ctx, db, session.User.ID, chi.URLParam(r, "id"))
	if err != nil {
		var nfErr domain.NotFoundError
		var fErr domain.ForbiddenError
		if errors.As(err, nfErr) {
			NotFoundHandler(w, r)
		} else if errors.As(err, fErr) {
			ForbiddenHandler(w, r)
		} else {
			InternalServerErrorHandler(w, r, err)
		}
		return
	}

	http.Redirect(w, r, BooksPath(pathUser.Username), http.StatusSeeOther)
}

func BookEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(queryExecer)
	pathUser := ctx.Value(RequestPathUserKey).(*minUser)

	bookID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		NotFoundHandler(w, r)
		return
	}

	uba := domain.UpdateBookArgs{}
	err = db.QueryRow(ctx, "select title, author, finish_date::text, media from books where id=$1 and user_id=$2", bookID, pathUser.ID).
		Scan(&uba.Title, &uba.Author, &uba.DateFinished, &uba.Media)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			NotFoundHandler(w, r)
		} else {
			InternalServerErrorHandler(w, r, err)
		}
		return
	}

	err = RenderBookEdit(w, baseViewDataFromRequest(r), bookID, uba, nil, pathUser.Username)
	if err != nil {
		InternalServerErrorHandler(w, r, err)
		return
	}
}

func BookUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(queryExecer)
	pathUser := ctx.Value(RequestPathUserKey).(*minUser)

	bookID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		NotFoundHandler(w, r)
		return
	}

	var found bool
	err = db.QueryRow(ctx, "select true from books where books.user_id=$1 and books.id=$2", pathUser.ID, bookID).Scan(&found)
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
			err := RenderBookEdit(w, baseViewDataFromRequest(r), bookID, uba, verr, pathUser.Username)
			if err != nil {
				InternalServerErrorHandler(w, r, err)
			}
			return
		}

		InternalServerErrorHandler(w, r, err)
		return
	}

	http.Redirect(w, r, BooksPath(pathUser.Username), http.StatusSeeOther)
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

	err = domain.ImportBooksFromCSV(ctx, db, pathUser.ID, file)
	if err != nil {
		InternalServerErrorHandler(w, r, err)
		return
	}

	http.Redirect(w, r, BooksPath(pathUser.Username), http.StatusSeeOther)
}
