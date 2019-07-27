package server

import (
	"net/http"
)

func RootHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := ctx.Value(RequestSessionKey).(*Session)

	if session.IsAuthenticated {
		http.Redirect(w, r, UserHomePath(session.User.Username), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, NewLoginPath(), http.StatusSeeOther)
	}
}
