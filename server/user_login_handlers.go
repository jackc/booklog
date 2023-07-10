package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/jackc/booklog/data"
	"github.com/jackc/booklog/route"
	"github.com/jackc/booklog/view"
	"github.com/jackc/errortree"
)

func UserLoginForm(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]any) error {
	var la data.UserLoginArgs
	return ctx.Value(RequestHTMLTemplateRendererKey).(*view.HTMLTemplateRenderer).ExecuteTemplate(w, "login.html", map[string]any{
		"bva":  baseViewArgsFromRequest(r),
		"form": la,
	})
}

func UserLogin(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]any) error {
	db := ctx.Value(RequestDBKey).(dbconn)

	la := data.UserLoginArgs{
		Username: r.FormValue("username"),
		Password: r.FormValue("password"),
	}

	userSessionID, err := data.UserLogin(ctx, db, la)
	if err != nil {
		var verr *errortree.Node
		if errors.As(err, &verr) {
			return ctx.Value(RequestHTMLTemplateRendererKey).(*view.HTMLTemplateRenderer).ExecuteTemplate(w, "login.html", map[string]any{
				"bva":  baseViewArgsFromRequest(r),
				"form": la,
				"verr": verr,
			})
		}

		return err
	}

	err = setSessionCookie(w, r, userSessionID)
	if err != nil {
		return err
	}

	http.Redirect(w, r, route.UserHomePath(la.Username), http.StatusSeeOther)
	return nil
}

func UserLogout(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]any) error {
	db := ctx.Value(RequestDBKey).(dbconn)
	session := ctx.Value(RequestSessionKey).(*Session)

	if session.IsAuthenticated {
		_, err := db.Exec(ctx, "delete from user_sessions where id=$1", session.ID)
		if err != nil {
			return err
		}
	}

	clearSessionCookie(w)

	http.Redirect(w, r, route.NewLoginPath(), http.StatusSeeOther)
	return nil
}
