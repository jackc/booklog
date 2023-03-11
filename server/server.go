package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/csrf"
	"github.com/gorilla/handlers"
	"github.com/gorilla/securecookie"
	"github.com/jackc/booklog/data"
	"github.com/jackc/booklog/route"
	"github.com/jackc/booklog/view"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

// Use when setting something though the request context.
type ctxRequestKey int

const (
	_ ctxRequestKey = iota
	RequestDBKey
	RequestSessionKey
	RequestPathUserKey
	RequestParamsKey
	RequestDevModeKey
	RequestHTMLTemplateRendererKey
)

type dbconn interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, optionsAndArgs ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, optionsAndArgs ...interface{}) pgx.Row
}

type Session struct {
	ID              [16]byte
	User            data.UserMin
	IsAuthenticated bool
	sc              *securecookie.SecureCookie
}

type AppServer struct {
	handler       http.Handler
	listenAddress string
	server        *http.Server
}

func NewAppServer(listenAddress string, csrfKey []byte, secureCookies bool, cookieHashKey []byte, cookieBlockKey []byte, dbpool *pgxpool.Pool, htr *view.HTMLTemplateRenderer) (*AppServer, error) {
	log := zerolog.New(os.Stdout).With().
		Timestamp().
		Logger()

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	r.Use(handlers.HTTPMethodOverrideHandler)

	r.Use(hlog.NewHandler(log))
	r.Use(hlog.RequestIDHandler("request_id", "x-request-id"))
	r.Use(hlog.MethodHandler("method"))
	r.Use(hlog.URLHandler("url"))
	r.Use(hlog.RemoteAddrHandler("remote_ip"))
	r.Use(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		hlog.FromRequest(r).Info().
			Int("status", status).
			Int("size", size).
			Dur("duration", duration).
			Msg("HTTP request")
	}))

	r.Use(middleware.Recoverer)

	CSRF := csrf.Protect(csrfKey, csrf.Secure(secureCookies))
	r.Use(CSRF)

	r.Use(devModeHandler(secureCookies))
	r.Use(pgxPoolHandler(dbpool))
	r.Use(htmlTemplateRendererHandler(htr))
	r.Use(parseParamsHandler())

	r.Use(sessionHandler(securecookie.New(cookieHashKey, cookieBlockKey)))

	r.Method("GET", "/", http.HandlerFunc(RootHandler))
	r.Method("GET", "/user_registration/new", http.HandlerFunc(UserRegistrationNew))
	r.Method("POST", "/user_registration", http.HandlerFunc(UserRegistrationCreate))

	r.Method("GET", "/login", http.HandlerFunc(UserLoginForm))
	r.Method("POST", "/login/handle", http.HandlerFunc(UserLogin))

	r.Method("POST", "/logout", http.HandlerFunc(UserLogout))

	r.Route("/users/{username}", func(r chi.Router) {
		r.Use(pathUserHandler())
		r.Use(requireSameSessionUserAndPathUserHandler())
		r.Method("GET", "/", http.HandlerFunc(UserHome))
		r.Method("GET", "/books", http.HandlerFunc(BookIndex))
		r.Method("GET", "/books/new", http.HandlerFunc(BookNew))
		r.Method("POST", "/books", http.HandlerFunc(BookCreate))
		r.Method("GET", "/books/{id}/edit", parseInt64URLParam("id")(http.HandlerFunc(BookEdit)))
		r.Method("GET", "/books/{id}", parseInt64URLParam("id")(http.HandlerFunc(BookShow)))
		r.Method("GET", "/books/{id}/confirm_delete", parseInt64URLParam("id")(http.HandlerFunc(BookConfirmDelete)))
		r.Method("PATCH", "/books/{id}", parseInt64URLParam("id")(http.HandlerFunc(BookUpdate)))
		r.Method("DELETE", "/books/{id}", parseInt64URLParam("id")(http.HandlerFunc(BookDelete)))
		r.Method("GET", "/books/import_csv/form", http.HandlerFunc(BookImportCSVForm))
		r.Method("POST", "/books/import_csv", http.HandlerFunc(BookImportCSV))
		r.Method("GET", "/books.csv", http.HandlerFunc(BookExportCSV))
	})

	fileServer(r, "/static", http.Dir("build/static"))

	return &AppServer{
		handler:       r,
		listenAddress: listenAddress,
	}, nil
}

func (s *AppServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}

func (s *AppServer) Serve() error {
	s.server = &http.Server{
		Addr:    s.listenAddress,
		Handler: s.handler,
	}

	fmt.Printf("Starting to listen on: %s\n", s.listenAddress)

	err := s.server.ListenAndServe()
	if err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (s *AppServer) Shutdown(ctx context.Context) error {
	s.server.SetKeepAlivesEnabled(false)
	err := s.server.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("graceful HTTP server shutdown failed: %w", err)
	}

	return nil
}

func fileServer(r chi.Router, path string, root http.FileSystem) {
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

func devModeHandler(devMode bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = context.WithValue(ctx, RequestDevModeKey, devMode)
			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}

func pgxPoolHandler(dbpool *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = context.WithValue(ctx, RequestDBKey, dbpool)
			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}

func htmlTemplateRendererHandler(htr *view.HTMLTemplateRenderer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = context.WithValue(ctx, RequestHTMLTemplateRendererKey, htr)
			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}

func sessionHandler(sc *securecookie.SecureCookie) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			session := &Session{sc: sc}
			ctx = context.WithValue(ctx, RequestSessionKey, session)

			cookie, err := r.Cookie("booklog-session-id")
			if err != nil {
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			var sessionID [16]byte
			err = sc.Decode("booklog-session-id", cookie.Value, &sessionID)
			if err != nil {
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			db := ctx.Value(RequestDBKey).(dbconn)
			err = db.QueryRow(ctx,
				"select user_sessions.id, users.id, users.username from user_sessions join users on user_sessions.user_id=users.id where user_sessions.id=$1",
				sessionID,
			).Scan(&session.ID, &session.User.ID, &session.User.Username)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					// invalid session ID
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				} else {
					InternalServerErrorHandler(w, r, err)
					return
				}
			}
			session.IsAuthenticated = true

			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}

func setSessionCookie(w http.ResponseWriter, r *http.Request, userSessionID [16]byte) error {
	ctx := r.Context()
	session := ctx.Value(RequestSessionKey).(*Session)

	encoded, err := session.sc.Encode("booklog-session-id", userSessionID)
	if err != nil {
		return err
	}

	cookie := &http.Cookie{
		Name:     "booklog-session-id",
		Value:    encoded,
		Path:     "/",
		Secure:   false, // TODO - true when not in insecure dev mode
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)

	return nil
}

func clearSessionCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     "booklog-session-id",
		Value:    "",
		Path:     "/",
		Secure:   false, // TODO - true when not in insecure dev mode
		HttpOnly: true,
		Expires:  time.Unix(0, 0),
	}
	http.SetCookie(w, cookie)
}

func pathUserHandler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			db := ctx.Value(RequestDBKey).(dbconn)

			user, err := data.GetUserMinByUsername(ctx, db, chi.URLParam(r, "username"))
			if err != nil {
				var nfErr *data.NotFoundError
				if errors.As(err, &nfErr) {
					NotFoundHandler(w, r)
				} else {
					InternalServerErrorHandler(w, r, err)
				}
				return
			}

			ctx = context.WithValue(ctx, RequestPathUserKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}

func requireSameSessionUserAndPathUserHandler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			session := ctx.Value(RequestSessionKey).(*Session)
			pathUser := ctx.Value(RequestPathUserKey).(*data.UserMin)

			if session.IsAuthenticated {
				if session.User.ID == pathUser.ID {
					next.ServeHTTP(w, r)
				} else {
					ForbiddenHandler(w, r)
				}
			} else {
				http.Redirect(w, r, route.NewLoginPath(), http.StatusSeeOther)
			}
		}

		return http.HandlerFunc(fn)
	}
}

type ctxURLParamKey string

func parseInt64URLParam(paramName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			n, err := strconv.ParseInt(chi.URLParam(r, paramName), 10, 64)
			if err != nil {
				NotFoundHandler(w, r)
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, ctxURLParamKey(paramName), n)
			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}

func int64URLParam(r *http.Request, name string) int64 {
	return r.Context().Value(ctxURLParamKey(name)).(int64)
}

func baseViewArgsFromRequest(r *http.Request) *view.BaseViewArgs {
	var pathUser *data.UserMin
	if um, ok := r.Context().Value(RequestPathUserKey).(*data.UserMin); ok {
		pathUser = um
	}

	session := r.Context().Value(RequestSessionKey).(*Session)
	var currentUser *data.UserMin
	if session.IsAuthenticated {
		currentUser = &session.User
	}

	var devMode bool
	if dm, ok := r.Context().Value(RequestDevModeKey).(bool); ok {
		devMode = dm
	}

	return &view.BaseViewArgs{
		CSRFField:   csrf.TemplateField(r),
		CurrentUser: currentUser,
		PathUser:    pathUser,
		DevMode:     devMode,
	}
}

func parseParamsHandler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			params, err := parseParams(r)
			if err != nil {
				InternalServerErrorHandler(w, r, err)
			}

			ctx = context.WithValue(ctx, RequestParamsKey, params)
			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}

func parseParams(r *http.Request) (map[string]any, error) {
	params := make(map[string]any)

	routeParams := chi.RouteContext(r.Context()).URLParams
	for i := 0; i < len(routeParams.Keys); i++ {
		params[routeParams.Keys[i]] = routeParams.Values[i]
	}

	err := r.ParseForm()
	if err != nil {
		return nil, err
	}

	for key, values := range r.Form {
		if len(values) > 0 {
			if strings.HasSuffix(key, "[]") {
				params[key[:len(key)-2]] = values
			} else {
				params[key] = values[0]
			}
		}
	}

	if r.Header.Get("Content-Type") == "application/json" {
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&params); err != nil {
			return nil, err
		}
	}

	return params, nil
}
