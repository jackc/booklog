package server

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/gorilla/csrf"
	"github.com/jackc/booklog/validate"
)

type BookUpdate struct {
}

func updateBook(ctx context.Context, db queryExecer, id int64, bcr *BookCreateRequest) error {
	v := validate.New()
	v.Presence("title", bcr.Title)
	v.Presence("author", bcr.Author)
	v.Presence("dateFinished", bcr.DateFinished)
	v.Presence("media", bcr.Media)

	if v.Err() != nil {
		return v.Err()
	}

	commandTag, err := db.Exec(ctx, "update finished_book set title=$1, author=$2, date_finished=$3, media=$4 where id=$5", bcr.Title, bcr.Author, bcr.DateFinished, bcr.Media, id)
	if err != nil {
		return err
	}
	if commandTag != "UPDATE 1" {
		return errors.New("not found")
	}

	return nil
}

func (action *BookUpdate) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(queryExecer)
	bookID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		panic(err)
	}

	var username string
	err = db.QueryRow(ctx, "select username from login_account join finished_book on login_account.id=finished_book.reader_id where finished_book.id=$1", bookID).Scan(&username)
	if err != nil {
		panic(err)
	}

	bcr := &BookCreateRequest{}
	bcr.Title = r.FormValue("title")
	bcr.Author = r.FormValue("author")
	bcr.DateFinished = r.FormValue("dateFinished")
	bcr.Media = r.FormValue("media")

	err = updateBook(ctx, db, bookID, bcr)
	if err != nil {
		// TODO - if errors is not a map this fails
		err := RenderBookEdit(w, csrf.TemplateField(r), bookID, bcr, err, username)
		if err != nil {
			panic(err)
		}
		return
	}

	http.Redirect(w, r, BooksPath(username), http.StatusSeeOther)

}
