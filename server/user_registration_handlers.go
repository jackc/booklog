package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/jackc/booklog/data"
	"github.com/jackc/booklog/route"
	"github.com/jackc/booklog/validate"
	"github.com/jackc/booklog/view"
)

func UserRegistrationNew(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var rua data.RegisterUserArgs
	return ctx.Value(RequestHTMLTemplateRendererKey).(*view.HTMLTemplateRenderer).ExecuteTemplate(w, "user_registration.html", map[string]any{
		"bva":  baseViewArgsFromRequest(r),
		"form": rua,
	})
}

func UserRegistrationCreate(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	db := ctx.Value(RequestDBKey).(dbconn)

	rua := data.RegisterUserArgs{
		Username: r.FormValue("username"),
		Password: r.FormValue("password"),
	}

	userSessionID, err := data.RegisterUser(ctx, db, rua)
	if err != nil {
		var verr validate.Errors
		if errors.As(err, &verr) {
			return ctx.Value(RequestHTMLTemplateRendererKey).(*view.HTMLTemplateRenderer).ExecuteTemplate(w, "user_registration.html", map[string]any{
				"bva":  baseViewArgsFromRequest(r),
				"form": rua,
				"verr": verr,
			})
		}

		return err
	}

	err = setSessionCookie(w, r, userSessionID)
	if err != nil {
		return err
	}

	http.Redirect(w, r, route.BooksPath(rua.Username), http.StatusSeeOther)
	return nil
}
