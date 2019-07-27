package server

import (
	"net/http"

	"github.com/jackc/booklog/route"
)

func RootHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := ctx.Value(RequestSessionKey).(*Session)

	if session.IsAuthenticated {
		http.Redirect(w, r, route.UserHomePath(session.User.Username), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, route.NewLoginPath(), http.StatusSeeOther)
	}
}
