package myhandler

import (
	"bytes"
	"context"
	"net/http"

	"github.com/jackc/booklog/view"
)

// Where does App / Server data go? How about per request data? Maybe via generics on Request? Maybe combine per server and per request data?

// Maybe Response is accessed via Request?

type Config[T any] struct {
	HTMLTemplateRenderer *view.HTMLTemplateRenderer

	BuildEnv   func(ctx context.Context, request *Request[T]) (*T, error)
	CleanupEnv func(ctx context.Context, request *Request[T]) error
}

func NewHandler[T any](config *Config[T], fn func(ctx context.Context, request *Request[T]) error) *Handler[T] {
	return &Handler[T]{
		fn:     fn,
		Config: config,
	}
}

type Handler[T any] struct {
	fn     func(ctx context.Context, request *Request[T]) error
	Config *Config[T]
}

func (h *Handler[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	request := &Request[T]{
		handler: h,

		request: r,

		responseWriter: w,
		buffer:         &bytes.Buffer{},
	}

	if h.Config.BuildEnv != nil {
		var err error
		request.Env, err = h.Config.BuildEnv(ctx, request)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
	if h.Config.CleanupEnv != nil {
		defer func() {
			_ = h.Config.CleanupEnv(ctx, request)
		}()
	}

	err := h.fn(ctx, request)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	_, _ = request.buffer.WriteTo(w)
	// TODO - what if write fails?
}

type Request[T any] struct {
	handler *Handler[T]
	Env     *T

	// Request specific fields.
	request *http.Request

	// Response specific fields.
	responseWriter http.ResponseWriter
	buffer         *bytes.Buffer
}

func (r *Request[T]) Request() *http.Request {
	return r.request
}

func (r *Request[T]) ResponseWriter() http.ResponseWriter {
	return r.responseWriter
}

func (r *Request[T]) RenderHTMLTemplate(name string, data any) error {
	return r.handler.Config.HTMLTemplateRenderer.ExecuteTemplate(r.buffer, name, data)
}
