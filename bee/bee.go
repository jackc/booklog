// Package bee provides a simple HTTP handler with functionality that is inconvenient to implement in middleware.
package bee

import (
	"bytes"
	"context"
	"net/http"
	"sync"
)

var bufPool = sync.Pool{
	New: func() any {
		return &bytes.Buffer{}
	},
}

type bufferedResponseWriter struct {
	w          http.ResponseWriter
	b          *bytes.Buffer
	statusCode int
}

func (brw *bufferedResponseWriter) Header() http.Header {
	return brw.w.Header()
}

func (brw *bufferedResponseWriter) Write(p []byte) (int, error) {
	return brw.b.Write(p)
}

func (brw *bufferedResponseWriter) WriteHeader(statusCode int) {
	brw.statusCode = statusCode
}

func (brw *bufferedResponseWriter) Reset() {
	brw.b.Reset()
}

type ErrorHandler func(w http.ResponseWriter, r *http.Request, err error) (bool, error)

// HandlerBuilder is used to build Handlers with shared functionality. HandlerBuilder must not be mutated after any
// methods have been called.
type HandlerBuilder struct {
	// ErrorHandlers are called one at a time until one returns true. If none return true or one returns an error then a
	// generic HTTP 500 error is returned.
	ErrorHandlers []ErrorHandler
}

func (hb *HandlerBuilder) New(fn func(ctx context.Context, w http.ResponseWriter, r *http.Request) error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := bufPool.Get().(*bytes.Buffer)
		defer func() {
			b.Reset()
			bufPool.Put(b)
		}()

		brw := &bufferedResponseWriter{
			w: w,
			b: b,
		}

		err := fn(r.Context(), brw, r)
		if err != nil {
			brw.Reset()
			for _, eh := range hb.ErrorHandlers {
				handled, err := eh(brw, r, err)
				if err != nil {
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
				if handled {
					break
				}
			}
		}

		if brw.statusCode != 0 {
			brw.w.WriteHeader(brw.statusCode)
		}
		brw.b.WriteTo(brw.w)
	})
}
