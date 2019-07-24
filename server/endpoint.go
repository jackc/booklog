package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog"
)

type Endpoint struct {
	DB                         *pgxpool.Pool
	Handler                    func(context.Context, *Endpoint, http.ResponseWriter, *http.Request)
	NotFoundHandler            func(context.Context, *Endpoint, http.ResponseWriter, *http.Request)
	InternalServerErrorHandler func(context.Context, *Endpoint, http.ResponseWriter, *http.Request, error)
}

func (e *Endpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	e.Handler(ctx, e, w, r)
}

func (e *Endpoint) InternalServerError(w http.ResponseWriter, r *http.Request, err error) {
	h := e.InternalServerErrorHandler
	if h == nil {
		h = DefaultInternalServerErrorHandler
	}
	h(r.Context(), e, w, r, err)
}

func (e *Endpoint) NotFound(w http.ResponseWriter, r *http.Request) {
	h := e.NotFoundHandler
	if h == nil {
		h = DefaultNotFoundHandler
	}
	h(r.Context(), e, w, r)
}

func DefaultInternalServerErrorHandler(ctx context.Context, e *Endpoint, w http.ResponseWriter, r *http.Request, err error) {
	zerolog.Ctx(ctx).Error().Err(err).Msg("internal server error")
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintln(w, err)
}

func DefaultNotFoundHandler(ctx context.Context, e *Endpoint, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintln(w, "Not found")
}
