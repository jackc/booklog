package server

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/gorilla/csrf"
)

type BookEdit struct {
}

func (action *BookEdit) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(queryExecer)

	bcr := &BookCreateRequest{}
	err := db.QueryRow(ctx, "select title, author, date_finished::text, media from book where id=$1", chi.URLParam(r, "id")).Scan(&bcr.Title, &bcr.Author, &bcr.DateFinished, &bcr.Media)
	// TODO - handle not found error
	// if len(result.Rows) == 0 {
	// 	http.NotFound(w, r)
	// 	return
	// }
	if err != nil {
		panic(err)
	}

	err = RenderBookEdit(w, csrf.TemplateField(r), chi.URLParam(r, "id"), bcr, map[string]string{})
	if err != nil {
		panic(err)
	}
}
