package data

import (
	"context"

	"github.com/jackc/booklog/validate"
	"github.com/jackc/pgx/v4"
	"golang.org/x/crypto/bcrypt"
	errors "golang.org/x/xerrors"
)

type UserLoginArgs struct {
	Username string
	Password string
}

func UserLogin(ctx context.Context, db dbconn, args UserLoginArgs) ([16]byte, error) {
	v := validate.New()
	v.Presence("username", args.Username)
	v.Presence("password", args.Password)

	if v.Err() != nil {
		return [16]byte{}, v.Err()
	}

	var userID int64
	var passwordDigest []byte

	err := db.QueryRow(ctx, "select id, password_digest from users where username=$1", args.Username).Scan(&userID, &passwordDigest)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			v.Add("base", errors.New("Invalid username or password."))
			return [16]byte{}, v.Err()
		}
		return [16]byte{}, err
	}

	err = bcrypt.CompareHashAndPassword(passwordDigest, []byte(args.Password))
	if err != nil {
		v.Add("base", errors.New("Invalid username or password."))
		return [16]byte{}, v.Err()
	}

	return createUserSession(ctx, db, userID)
}
