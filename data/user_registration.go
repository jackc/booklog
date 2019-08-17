package data

import (
	"context"

	"github.com/jackc/booklog/validate"
	"golang.org/x/crypto/bcrypt"
)

type RegisterUserArgs struct {
	Username string
	Password string
}

func RegisterUser(ctx context.Context, db dbconn, args RegisterUserArgs) ([16]byte, error) {
	v := validate.New()
	v.Presence("username", args.Username)
	v.Presence("password", args.Password)
	v.MinLength("password", args.Password, 8)

	if v.Err() != nil {
		return [16]byte{}, v.Err()
	}

	passwordDigest, err := bcrypt.GenerateFromPassword([]byte(args.Password), bcrypt.DefaultCost)
	if err != nil {
		return [16]byte{}, err
	}

	var userID int64
	err = db.QueryRow(ctx, "insert into users(username, password_digest) values($1, $2) returning id", args.Username, passwordDigest).Scan(&userID)
	if err != nil {
		return [16]byte{}, err
	}

	return createUserSession(ctx, db, userID)
}
