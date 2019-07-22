package server

import (
	"net/http"
	"time"

	"github.com/jackc/booklog/domain"
	"github.com/jackc/booklog/validate"
	errors "golang.org/x/xerrors"
)

func UserLoginForm(w http.ResponseWriter, r *http.Request) {
	var la domain.UserLoginArgs

	err := RenderUserLoginForm(w, baseViewDataFromRequest(r), la, nil)
	if err != nil {
		panic(err)
	}
}

func UserLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(queryExecer)

	la := domain.UserLoginArgs{
		Username: r.FormValue("username"),
		Password: r.FormValue("password"),
	}

	userSessionID, err := domain.UserLogin(ctx, db, la)
	if err != nil {
		var verr validate.Errors
		if errors.As(err, &verr) {
			err := RenderUserLoginForm(w, baseViewDataFromRequest(r), la, verr)
			if err != nil {
				panic(err)
			}

			return
		}

		panic(err)
	}

	encoded, err := sc.Encode("booklog-session-id", userSessionID)
	if err != nil {
		panic(err)
	}

	cookie := &http.Cookie{
		Name:     "booklog-session-id",
		Value:    encoded,
		Path:     "/",
		Secure:   false, // TODO - true when not in insecure dev mode
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)

	http.Redirect(w, r, BooksPath(la.Username), http.StatusSeeOther)
}

func UserLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(queryExecer)
	session := ctx.Value(RequestSessionKey).(*Session)

	if session.IsAuthenticated {
		_, err := db.Exec(ctx, "delete from user_session where id=$1", session.ID)
		if err != nil {
			panic(err)
		}
	}

	cookie := &http.Cookie{
		Name:     "booklog-session-id",
		Value:    "",
		Path:     "/",
		Secure:   false, // TODO - true when not in insecure dev mode
		HttpOnly: true,
		Expires:  time.Unix(0, 0),
	}
	http.SetCookie(w, cookie)

	http.Redirect(w, r, NewLoginPath(), http.StatusSeeOther)
}
