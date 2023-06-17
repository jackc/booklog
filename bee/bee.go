// Package bee provides a simple HTTP handler with functionality that is inconvenient to implement in middleware.
//
// It provides two primary features. First, is easier error handling. Handlers can return errors which will be handled
// by a list of error handlers that will be called when an error occurs. Second, it automatically sets the ETag header
// based on the digest of the response body.
//
// These features may seem entirely unrelated but they are both related because the response body must be buffered in
// its entirety. For error handling an error may occur after some of the response has been written and the response
// needs to be replaced. For ETag the response body must be buffered so that the digest can be calculated and set in the
// headers.
package bee

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"regexp"
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

	// ETagDigestFilter is used to filter out parts of the response body that should not be included in the automatic ETag
	// digest. This is useful for filtering out dynamic content such as CSRF tokens. If nil then the entire response body
	// is used.
	ETagDigestFilter *regexp.Regexp
}

// New returns a new http.Handler that calls fn. If fn returns an error then the error is passed to the ErrorHandlers.
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

		// Even though the net/http package will set the Content-Type header if it is not set, we do it here so that
		// Content-Type is available for middleware such as chi/middleware/Compress.
		if brw.Header().Get("Content-Type") == "" {
			brw.Header().Set("Content-Type", http.DetectContentType(brw.b.Bytes()))
		}

		if r.Method == http.MethodGet && brw.Header().Get("ETag") == "" {
			digest := sha256.New()
			if hb.ETagDigestFilter == nil {
				digest.Write(brw.b.Bytes())
			} else {
				buf := brw.b.Bytes()
				pos := 0
				for pos < len(buf) {
					loc := hb.ETagDigestFilter.FindIndex(buf[pos:])
					if loc == nil {
						digest.Write(buf[pos:])
						pos = len(buf)
					} else {
						digest.Write(buf[pos : pos+loc[0]])
						pos = loc[1]
					}
				}
			}

			bodyDigest := digest.Sum(nil)
			etag := `W/"` + base64.URLEncoding.EncodeToString(bodyDigest[:]) + `"`

			if r.Header.Get("If-None-Match") == etag {
				brw.w.WriteHeader(http.StatusNotModified)
				return
			}

			brw.w.Header().Set("ETag", etag)
		}

		if brw.statusCode != 0 {
			brw.w.WriteHeader(brw.statusCode)
		}
		brw.b.WriteTo(brw.w)
	})
}
