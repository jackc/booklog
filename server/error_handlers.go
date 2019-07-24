package server

import (
	"fmt"
	"net/http"

	"github.com/rs/zerolog/hlog"
)

func InternalServerErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	hlog.FromRequest(r).Error().Err(err).Msg("internal server error")
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintln(w, "Internal server error")
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintln(w, "Not found")
}
