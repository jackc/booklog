package server

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/gorilla/csrf"
)

type BookEdit struct {
}

func (action *BookEdit) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(queryExecer)
	username := chi.URLParam(r, "username")
	bookID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		panic(err)
	}

	bcr := &BookCreateRequest{}
	err = db.QueryRow(ctx, "select title, author, date_finished::text, media from finished_book where id=$1", bookID).Scan(&bcr.Title, &bcr.Author, &bcr.DateFinished, &bcr.Media)
	// TODO - handle not found error
	// if len(result.Rows) == 0 {
	// 	http.NotFound(w, r)
	// 	return
	// }
	if err != nil {
		panic(err)
	}

	err = RenderBookEdit(w, csrf.TemplateField(r), bookID, bcr, map[string]string{}, username)
	if err != nil {
		panic(err)
	}
}
