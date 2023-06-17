package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/jackc/booklog/data"
	"github.com/jackc/booklog/myhandler"
	"github.com/jackc/booklog/route"
	"github.com/jackc/booklog/validate"
)

func UserRegistrationNew(ctx context.Context, request *myhandler.Request[HandlerEnv]) error {
	var rua data.RegisterUserArgs
	return request.RenderHTMLTemplate("user_registration.html", map[string]any{
		"bva":  baseViewArgsFromRequest(request),
		"form": rua,
	})
}

func UserRegistrationCreate(ctx context.Context, request *myhandler.Request[HandlerEnv]) error {
	db := request.Env.dbconn

	rua := data.RegisterUserArgs{
		Username: request.Request().FormValue("username"),
		Password: request.Request().FormValue("password"),
	}

	userSessionID, err := data.RegisterUser(ctx, db, rua)
	if err != nil {
		var verr validate.Errors
		if errors.As(err, &verr) {
			return request.RenderHTMLTemplate("user_registration.html", map[string]any{
				"bva":  baseViewArgsFromRequest(request),
				"form": rua,
				"verr": verr,
			})
		}

		return err
	}

	err = setSessionCookie(request.ResponseWriter(), request.Request(), userSessionID)
	if err != nil {
		return err
	}

	http.Redirect(request.ResponseWriter(), request.Request(), route.BooksPath(rua.Username), http.StatusSeeOther)
	return nil
}
