package server

import (
	"net/http"

	"github.com/jackc/booklog/data"
	"github.com/jackc/booklog/validate"
	errors "golang.org/x/xerrors"
)

func UserRegistrationNew(w http.ResponseWriter, r *http.Request) {
	var rua data.RegisterUserArgs

	err := RenderUserRegistrationNew(w, baseViewDataFromRequest(r), rua, nil)
	if err != nil {
		InternalServerErrorHandler(w, r, err)
		return
	}
}

func UserRegistrationCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(queryExecer)

	rua := data.RegisterUserArgs{
		Username: r.FormValue("username"),
		Password: r.FormValue("password"),
	}

	userSessionID, err := data.RegisterUser(ctx, db, rua)
	if err != nil {
		var verr validate.Errors
		if errors.As(err, &verr) {
			err := RenderUserRegistrationNew(w, baseViewDataFromRequest(r), rua, verr)
			if err != nil {
				InternalServerErrorHandler(w, r, err)
			}
			return
		}

		InternalServerErrorHandler(w, r, err)
		return
	}

	err = setSessionCookie(w, r, userSessionID)
	if err != nil {
		InternalServerErrorHandler(w, r, err)
		return
	}

	http.Redirect(w, r, BooksPath(rua.Username), http.StatusSeeOther)
}
