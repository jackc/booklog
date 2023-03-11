package server

import (
	"errors"
	"net/http"

	"github.com/jackc/booklog/data"
	"github.com/jackc/booklog/route"
	"github.com/jackc/booklog/validate"
	"github.com/jackc/booklog/view"
)

func UserRegistrationNew(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	htr := ctx.Value(RequestHTMLTemplateRendererKey).(*view.HTMLTemplateRenderer)
	var rua data.RegisterUserArgs

	err := htr.ExecuteTemplate(w, "user_registration.html", map[string]any{
		"bva":  baseViewArgsFromRequest(r),
		"form": rua,
	})
	if err != nil {
		InternalServerErrorHandler(w, r, err)
		return
	}
}

func UserRegistrationCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(dbconn)
	htr := ctx.Value(RequestHTMLTemplateRendererKey).(*view.HTMLTemplateRenderer)

	rua := data.RegisterUserArgs{
		Username: r.FormValue("username"),
		Password: r.FormValue("password"),
	}

	userSessionID, err := data.RegisterUser(ctx, db, rua)
	if err != nil {
		var verr validate.Errors
		if errors.As(err, &verr) {
			err := htr.ExecuteTemplate(w, "user_registration.html", map[string]any{
				"bva":  baseViewArgsFromRequest(r),
				"form": rua,
				"verr": verr,
			})
			if err != nil {
				InternalServerErrorHandler(w, r, err)
			}
			return
		}

		InternalServerErrorHandler(w, r, err)
		return
	}

	err = setSessionCookie(w, r, userSessionID)
	if err != nil {
		InternalServerErrorHandler(w, r, err)
		return
	}

	http.Redirect(w, r, route.BooksPath(rua.Username), http.StatusSeeOther)
}
