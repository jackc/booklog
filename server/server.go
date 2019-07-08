package server

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gorilla/csrf"
	"github.com/gorilla/handlers"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"github.com/spf13/viper"
)

// Key to use when setting the DB.
type ctxKeyDB int

// RequestDBKey is the key that holds the DB for this request.
const RequestDBKey ctxKeyDB = 0

type queryExecer interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, optionsAndArgs ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, optionsAndArgs ...interface{}) pgx.Row
}

func Serve(listenAddress string, csrfKey []byte, insecureDevMode bool) {
	log := zerolog.New(os.Stdout).With().
		Timestamp().
		Logger()

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	r.Use(handlers.HTTPMethodOverrideHandler)

	r.Use(middleware.Logger)

	r.Use(hlog.NewHandler(log))
	r.Use(hlog.URLHandler("url"))

	r.Use(middleware.Recoverer)

	CSRF := csrf.Protect(csrfKey, csrf.Secure(!insecureDevMode))
	r.Use(CSRF)

	err := LoadTemplates("html")
	if err != nil {
		panic(err)
	}

	dbpool, err := pool.Connect(context.Background(), viper.GetString("database_uri"))
	if err != nil {
		panic(err)
	}
	r.Use(func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = context.WithValue(ctx, RequestDBKey, dbpool)
			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	})

	r.Method("GET", "/", &BookIndex{})
	r.Method("GET", "/books", &BookIndex{})
	r.Method("GET", "/books/new", &BookNew{})
	r.Method("POST", "/books", &BookCreate{})
	r.Method("GET", "/books/{id}/edit", &BookEdit{})
	r.Method("PATCH", "/books/{id}", &BookUpdate{})
	r.Method("DELETE", "/books/{id}", &BookDelete{})

	fileServer(r, "/static", http.Dir("build/static"))

	http.ListenAndServe(listenAddress, r)
}

func fileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	fs := http.StripPrefix(path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}
