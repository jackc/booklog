package server

import (
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

func Serve(listenAddress string) {
	log := zerolog.New(os.Stdout).With().
		Timestamp().
		Logger()

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)

	r.Use(hlog.NewHandler(log))
	r.Use(hlog.URLHandler("url"))

	r.Use(middleware.Recoverer)

	r.Method("GET", "/", &BookIndex{})
	r.Method("GET", "/new", &BookNew{})
	r.Method("POST", "/books", &BookCreate{})
	r.Method("POST", "/books/{id}/delete", &BookDelete{})
	http.ListenAndServe(listenAddress, r)
}
