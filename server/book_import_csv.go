package server

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/jackc/booklog/domain"
)

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
