package server

import (
	"net/http"

	"github.com/jackc/booklog/data"
	"github.com/jackc/booklog/validate"
	errors "golang.org/x/xerrors"
)

func UserLoginForm(w http.ResponseWriter, r *http.Request) {
	var la data.UserLoginArgs

	err := RenderUserLoginForm(w, baseViewDataFromRequest(r), la, nil)
	if err != nil {
		InternalServerErrorHandler(w, r, err)
		return
	}
}

func UserLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(queryExecer)

	la := data.UserLoginArgs{
		Username: r.FormValue("username"),
		Password: r.FormValue("password"),
	}

	userSessionID, err := data.UserLogin(ctx, db, la)
	if err != nil {
		var verr validate.Errors
		if errors.As(err, &verr) {
			err := RenderUserLoginForm(w, baseViewDataFromRequest(r), la, verr)
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

	http.Redirect(w, r, UserHomePath(la.Username), http.StatusSeeOther)
}

func UserLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(queryExecer)
	session := ctx.Value(RequestSessionKey).(*Session)

	if session.IsAuthenticated {
		_, err := db.Exec(ctx, "delete from user_sessions where id=$1", session.ID)
		if err != nil {
			InternalServerErrorHandler(w, r, err)
			return
		}
	}

	clearSessionCookie(w)

	http.Redirect(w, r, NewLoginPath(), http.StatusSeeOther)
}
