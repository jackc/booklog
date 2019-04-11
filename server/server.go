package server

import (
	"context"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"

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

	templates, err := loadTemplates()
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

	r.Method("GET", "/", &BookIndex{templates: templates})
	r.Method("GET", "/books", &BookIndex{templates: templates})
	r.Method("GET", "/books/new", &BookNew{templates: templates})
	r.Method("POST", "/books", &BookCreate{templates: templates})
	r.Method("GET", "/books/{id}/edit", &BookEdit{templates: templates})
	r.Method("PATCH", "/books/{id}", &BookUpdate{templates: templates})
	r.Method("DELETE", "/books/{id}", &BookDelete{})
	http.ListenAndServe(listenAddress, r)
}

func loadTemplates() (*template.Template, error) {
	root := template.New("root")
	root.Funcs(template.FuncMap{
		"newBookPath":  NewBookPath,
		"bookPath":     BookPath,
		"editBookPath": EditBookPath,
		"booksPath":    BooksPath,
	})

	targets := []struct {
		name     string
		filepath string
	}{
		{name: "book_index", filepath: "html/book_index.html"},
		{name: "book_new", filepath: "html/book_new.html"},
		{name: "book_edit", filepath: "html/book_edit.html"},
	}

	for _, t := range targets {
		src, err := ioutil.ReadFile(t.filepath)
		if err != nil {
			return nil, err
		}

		tmpl := root.New(t.name)
		tmpl, err = tmpl.Parse(string(src))
		if err != nil {
			return nil, err
		}
	}

	return root, nil
}
