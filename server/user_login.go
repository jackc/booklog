package server

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/booklog/domain"
	"github.com/jackc/booklog/validate"
	errors "golang.org/x/xerrors"
)

func UserLoginForm(ctx context.Context, e *Endpoint, w http.ResponseWriter, r *http.Request) {
	var la domain.UserLoginArgs

	err := RenderUserLoginForm(w, baseViewDataFromRequest(r), la, nil)
	if err != nil {
		e.InternalServerError(w, r, err)
		return
	}
}

func UserLogin(ctx context.Context, e *Endpoint, w http.ResponseWriter, r *http.Request) {
	la := domain.UserLoginArgs{
		Username: r.FormValue("username"),
		Password: r.FormValue("password"),
	}

	userSessionID, err := domain.UserLogin(ctx, e.DB, la)
	if err != nil {
		var verr validate.Errors
		if errors.As(err, &verr) {
			err := RenderUserLoginForm(w, baseViewDataFromRequest(r), la, verr)
			if err != nil {
				e.InternalServerError(w, r, err)
			}
			return
		}

		e.InternalServerError(w, r, err)
		return
	}

	encoded, err := sc.Encode("booklog-session-id", userSessionID)
	if err != nil {
		e.InternalServerError(w, r, err)
		return
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

func UserLogout(ctx context.Context, e *Endpoint, w http.ResponseWriter, r *http.Request) {
	session := ctx.Value(RequestSessionKey).(*Session)

	if session.IsAuthenticated {
		_, err := e.DB.Exec(ctx, "delete from user_sessions where id=$1", session.ID)
		if err != nil {
			e.InternalServerError(w, r, err)
			return
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
