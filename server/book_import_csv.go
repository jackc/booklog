package server

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/jackc/booklog/domain"
)

func BookImportCSVForm(ctx context.Context, e *Endpoint, w http.ResponseWriter, r *http.Request) {
	err := RenderBookImportCSVForm(w, baseViewDataFromRequest(r), chi.URLParam(r, "username"))
	if err != nil {
		e.InternalServerError(w, r, err)
		return
	}
}

func BookImportCSV(ctx context.Context, e *Endpoint, w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	var userID int64
	err := e.DB.QueryRow(ctx, "select id from users where username=$1", username).Scan(&userID)
	if err != nil {
		e.NotFound(w, r)
		return
	}

	r.ParseMultipartForm(10 << 20)

	file, _, err := r.FormFile("file")
	if err != nil {
		e.InternalServerError(w, r, err)
		return
	}
	defer file.Close()

	err = domain.ImportBooksFromCSV(ctx, e.DB, userID, file)
	if err != nil {
		e.InternalServerError(w, r, err)
		return
	}

	http.Redirect(w, r, BooksPath(username), http.StatusSeeOther)
}
