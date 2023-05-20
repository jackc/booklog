package server

import (
	"context"
	"net/http"

	"github.com/jackc/booklog/myhandler"
	"github.com/jackc/booklog/route"
)

func RootHandler(ctx context.Context, request *myhandler.Request[HandlerEnv]) error {
	session := ctx.Value(RequestSessionKey).(*Session)

	if session.IsAuthenticated {
		http.Redirect(request.ResponseWriter(), request.Request(), route.UserHomePath(session.User.Username), http.StatusSeeOther)
	} else {
		http.Redirect(request.ResponseWriter(), request.Request(), route.NewLoginPath(), http.StatusSeeOther)
	}

	return nil
}
