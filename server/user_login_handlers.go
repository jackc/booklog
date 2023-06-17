package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/jackc/booklog/data"
	"github.com/jackc/booklog/myhandler"
	"github.com/jackc/booklog/route"
	"github.com/jackc/booklog/validate"
	"github.com/jackc/booklog/view"
)

func UserLoginForm(ctx context.Context, request *myhandler.Request[HandlerEnv]) error {
	var la data.UserLoginArgs
	return ctx.Value(RequestHTMLTemplateRendererKey).(*view.HTMLTemplateRenderer).ExecuteTemplate(request.ResponseWriter(), "login.html", map[string]any{
		"bva":  baseViewArgsFromRequest(request.Request()),
		"form": la,
	})
}

func UserLogin(ctx context.Context, request *myhandler.Request[HandlerEnv]) error {
	db := ctx.Value(RequestDBKey).(dbconn)

	la := data.UserLoginArgs{
		Username: request.Request().FormValue("username"),
		Password: request.Request().FormValue("password"),
	}

	userSessionID, err := data.UserLogin(ctx, db, la)
	if err != nil {
		var verr validate.Errors
		if errors.As(err, &verr) {
			return ctx.Value(RequestHTMLTemplateRendererKey).(*view.HTMLTemplateRenderer).ExecuteTemplate(request.ResponseWriter(), "login.html", map[string]any{
				"bva":  baseViewArgsFromRequest(request.Request()),
				"form": la,
				"verr": verr,
			})
		}

		return err
	}

	err = setSessionCookie(request.ResponseWriter(), request.Request(), userSessionID)
	if err != nil {
		return err
	}

	http.Redirect(request.ResponseWriter(), request.Request(), route.UserHomePath(la.Username), http.StatusSeeOther)
	return nil
}

func UserLogout(ctx context.Context, request *myhandler.Request[HandlerEnv]) error {
	db := ctx.Value(RequestDBKey).(dbconn)
	session := ctx.Value(RequestSessionKey).(*Session)

	if session.IsAuthenticated {
		_, err := db.Exec(ctx, "delete from user_sessions where id=$1", session.ID)
		if err != nil {
			return err
		}
	}

	clearSessionCookie(request.ResponseWriter())

	http.Redirect(request.ResponseWriter(), request.Request(), route.NewLoginPath(), http.StatusSeeOther)
	return nil
}
