package server

import (
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

func (f BookEditForm) Parse() (ParsedBookEditForm, validate.Errors) {
	var err error
	p := ParsedBookEditForm{BookEditForm: f}
	v := validate.New()

	p.DateFinished, err = time.Parse("2006-01-02", f.DateFinished)
	if err != nil {
		v.Add("dateFinished", errors.New("is not a date"))
	}

	if v.Err() != nil {
		return p, v.Err().(validate.Errors)
	}

	return p, nil
}

type ParsedBookEditForm struct {
	BookEditForm
	DateFinished time.Time
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
	parsedForm, verr := form.Parse()
	if verr != nil {
		err := RenderBookNew(w, baseViewDataFromRequest(r), form, verr, pathUser.Username)
		if err != nil {
			InternalServerErrorHandler(w, r, err)
		}
		return
	}

	cba := domain.CreateBookArgs{
		UserID:       pathUser.ID,
		Title:        parsedForm.Title,
		Author:       parsedForm.Author,
		DateFinished: parsedForm.DateFinished,
		Media:        parsedForm.Media,
	}

	err := domain.CreateBook(ctx, db, cba)
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

	http.Redirect(w, r, BooksPath(pathUser.Username), http.StatusSeeOther)
}

func BookDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(queryExecer)
	session := ctx.Value(RequestSessionKey).(*Session)
	pathUser := ctx.Value(RequestPathUserKey).(*minUser)
	bookID := int64URLParam(r, "id")

	err := domain.DeleteBook(ctx, db, session.User.ID, bookID)
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
	session := ctx.Value(RequestSessionKey).(*Session)
	bookID := int64URLParam(r, "id")

	form := BookEditForm{
		Title:        r.FormValue("title"),
		Author:       r.FormValue("author"),
		DateFinished: r.FormValue("dateFinished"),
		Media:        r.FormValue("media"),
	}
	parsedForm, verr := form.Parse()
	if verr != nil {
		err := RenderBookEdit(w, baseViewDataFromRequest(r), bookID, form, verr, pathUser.Username)
		if err != nil {
			InternalServerErrorHandler(w, r, err)
		}
		return
	}

	uba := domain.UpdateBookArgs{
		Title:        parsedForm.Title,
		Author:       parsedForm.Author,
		DateFinished: parsedForm.DateFinished,
		Media:        parsedForm.Media,
	}

	err := domain.UpdateBook(ctx, db, session.User.ID, bookID, uba)
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
