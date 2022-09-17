package data

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type UserMin struct {
	ID       int64
	Username string
}

func GetUserMinByUsername(ctx context.Context, db dbconn, username string) (*UserMin, error) {
	var user UserMin
	err := db.QueryRow(ctx, "select id, username from users where username=$1", username).Scan(&user.ID, &user.Username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &NotFoundError{target: fmt.Sprintf("user username=%s", username)}
		}
		return nil, err
	}

	return &user, nil
}
