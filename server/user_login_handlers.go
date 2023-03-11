package server

import (
	"errors"
	"net/http"

	"github.com/jackc/booklog/data"
	"github.com/jackc/booklog/route"
	"github.com/jackc/booklog/validate"
	"github.com/jackc/booklog/view"
)

func UserLoginForm(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	htr := ctx.Value(RequestHTMLTemplateRendererKey).(*view.HTMLTemplateRenderer)
	var la data.UserLoginArgs
	err := htr.ExecuteTemplate(w, "login.html", map[string]any{
		"bva":  baseViewArgsFromRequest(r),
		"form": la,
	})
	if err != nil {
		InternalServerErrorHandler(w, r, err)
	}
}

func UserLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(dbconn)
	htr := ctx.Value(RequestHTMLTemplateRendererKey).(*view.HTMLTemplateRenderer)

	la := data.UserLoginArgs{
		Username: r.FormValue("username"),
		Password: r.FormValue("password"),
	}

	userSessionID, err := data.UserLogin(ctx, db, la)
	if err != nil {
		var verr validate.Errors
		if errors.As(err, &verr) {
			err := htr.ExecuteTemplate(w, "login.html", map[string]any{
				"bva":  baseViewArgsFromRequest(r),
				"form": la,
				"verr": verr,
			})
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

	http.Redirect(w, r, route.UserHomePath(la.Username), http.StatusSeeOther)
}

func UserLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(dbconn)
	session := ctx.Value(RequestSessionKey).(*Session)

	if session.IsAuthenticated {
		_, err := db.Exec(ctx, "delete from user_sessions where id=$1", session.ID)
		if err != nil {
			InternalServerErrorHandler(w, r, err)
			return
		}
	}

	clearSessionCookie(w)

	http.Redirect(w, r, route.NewLoginPath(), http.StatusSeeOther)
}
