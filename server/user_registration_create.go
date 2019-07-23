package server

import (
	"context"
	"net/http"

	"github.com/jackc/booklog/domain"
	"github.com/jackc/booklog/validate"
	errors "golang.org/x/xerrors"
)

func UserRegistrationCreate(ctx context.Context, e *Endpoint, w http.ResponseWriter, r *http.Request) {
	rua := domain.RegisterUserArgs{
		Username: r.FormValue("username"),
		Password: r.FormValue("password"),
	}

	err := domain.RegisterUser(ctx, e.DB, rua)
	if err != nil {
		var verr validate.Errors
		if errors.As(err, &verr) {
			err := RenderUserRegistrationNew(w, baseViewDataFromRequest(r), rua, verr)
			if err != nil {
				e.InternalServerError(w, r, err)
			}
			return
		}

		e.InternalServerError(w, r, err)
		return
	}

	http.Redirect(w, r, BooksPath(rua.Username), http.StatusSeeOther)
}
