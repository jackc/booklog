package server

import (
	"context"
	"net/http"

	"github.com/jackc/booklog/route"
)

func RootHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]any) error {
	session := ctx.Value(RequestSessionKey).(*Session)

	if session.IsAuthenticated {
		http.Redirect(w, r, route.UserHomePath(session.User.Username), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, route.NewLoginPath(), http.StatusSeeOther)
	}

	return nil
}
