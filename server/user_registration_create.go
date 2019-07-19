package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/jackc/booklog/validate"
	"golang.org/x/crypto/bcrypt"
)

func registerUser(ctx context.Context, db queryExecer, urr *UserRegistrationRequest) error {
	v := validate.New()
	v.Presence("username", urr.Username)
	v.Presence("password", urr.Password)
	v.MinLength("password", urr.Password, 8)

	if v.Err() != nil {
		return v.Err()
	}

	passwordDigest, err := bcrypt.GenerateFromPassword([]byte(urr.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, "insert into login_account(username, password_digest) values($1, $2)", urr.Username, passwordDigest)
	if err != nil {
		return err
	}

	return nil
}

func UserRegistrationCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := ctx.Value(RequestDBKey).(queryExecer)

	urr := &UserRegistrationRequest{}
	urr.Username = r.FormValue("username")
	urr.Password = r.FormValue("password")

	err := registerUser(ctx, db, urr)
	if err != nil {
		fmt.Println(err)
		err := RenderUserRegistrationNew(w, csrf.TemplateField(r), urr, err)
		if err != nil {
			panic(err)
		}

		if err != nil {
			panic(err)
		}
		return
	}

	http.Redirect(w, r, BooksPath(urr.Username), http.StatusSeeOther)
}
