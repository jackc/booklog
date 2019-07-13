package server

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/gorilla/csrf"
	"github.com/jackc/booklog/validate"
)

type BookCreate struct {
}

func createBook(ctx context.Context, db queryExecer, bcr *BookCreateRequest, readerID int64) error {
	v := validate.New()
	v.Presence("title", bcr.Title)
	v.Presence("author", bcr.Author)
	v.Presence("dateFinished", bcr.DateFinished)
	v.Presence("media", bcr.Media)

	if v.Err() != nil {
		return v.Err()
	}

	_, err := db.Exec(ctx, "insert into finished_book(reader_id, title, author, date_finished, media) values($1, $2, $3, $4, $5)", readerID, bcr.Title, bcr.Author, bcr.DateFinished, bcr.Media)
	if err != nil {
		return err
	}

	return nil
}

func (action *BookCreate) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(queryExecer)
	username := chi.URLParam(r, "username")
	var readerID int64
	err := db.QueryRow(ctx, "select id from login_account where username=$1", username).Scan(&readerID)
	if err != nil {
		panic(err)
	}

	bcr := &BookCreateRequest{}
	bcr.Title = r.FormValue("title")
	bcr.Author = r.FormValue("author")
	bcr.DateFinished = r.FormValue("dateFinished")
	bcr.Media = r.FormValue("media")

	err = createBook(ctx, db, bcr, readerID)
	if err != nil {
		err := RenderBookNew(w, csrf.TemplateField(r), bcr, err, username)
		if err != nil {
			panic(err)
		}

		if err != nil {
			panic(err)
		}
		return
	}

	http.Redirect(w, r, BooksPath(username), http.StatusSeeOther)

}
