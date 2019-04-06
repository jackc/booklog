package server

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gorilla/csrf"
	"github.com/gorilla/handlers"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

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

	r.Method("GET", "/", &BookIndex{templates: templates})
	r.Method("GET", "/books", &BookIndex{templates: templates})
	r.Method("GET", "/books/new", &BookNew{templates: templates})
	r.Method("POST", "/books", &BookCreate{templates: templates})
	r.Method("DELETE", "/books/{id}", &BookDelete{})
	http.ListenAndServe(listenAddress, r)
}

func loadTemplates() (*template.Template, error) {
	root := template.New("root")
	root.Funcs(template.FuncMap{
		"newBookPath": NewBookPath,
		"bookPath":    BookPath,
		"booksPath":   BooksPath,
	})

	targets := []struct {
		name     string
		filepath string
	}{
		{name: "book_index", filepath: "html/book_index.html"},
		{name: "book_new", filepath: "html/book_new.html"},
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
