package server

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/jackc/booklog/domain"
)

func BookImportCSVForm(w http.ResponseWriter, r *http.Request) {
	err := RenderBookImportCSVForm(w, baseViewDataFromRequest(r), chi.URLParam(r, "username"))
	if err != nil {
		panic(err)
	}
}

func BookImportCSV(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(queryExecer)

	username := chi.URLParam(r, "username")
	var readerID int64
	err := db.QueryRow(ctx, "select id from users where username=$1", username).Scan(&readerID)
	if err != nil {
		panic(err)
	}

	r.ParseMultipartForm(10 << 20)

	file, _, err := r.FormFile("file")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	domain.ImportBooksFromCSV(ctx, db, readerID, file)

	http.Redirect(w, r, BooksPath(username), http.StatusSeeOther)
}
