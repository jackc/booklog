package domain

import (
	"context"

	"github.com/jackc/booklog/validate"
	"golang.org/x/crypto/bcrypt"
)

type RegisterUserArgs struct {
	Username string
	Password string
}

func RegisterUser(ctx context.Context, db queryExecer, args RegisterUserArgs) error {
	v := validate.New()
	v.Presence("username", args.Username)
	v.Presence("password", args.Password)
	v.MinLength("password", args.Password, 8)

	if v.Err() != nil {
		return v.Err()
	}

	passwordDigest, err := bcrypt.GenerateFromPassword([]byte(args.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, "insert into login_account(username, password_digest) values($1, $2)", args.Username, passwordDigest)
	if err != nil {
		return err
	}

	return nil
}
