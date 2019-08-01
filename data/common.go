package data

import (
	"context"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

type queryExecer interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, optionsAndArgs ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, optionsAndArgs ...interface{}) pgx.Row
}

type scanner interface {
	Scan(...interface{}) error
}

func createUserSession(ctx context.Context, db queryExecer, userID int64) ([16]byte, error) {
	var userSessionID [16]byte
	err := db.QueryRow(ctx, "insert into user_sessions(user_id) values ($1) returning id", userID).Scan(&userSessionID)
	return userSessionID, err
}

type NotFoundError struct {
	target string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("not found: %s", e.target)
}
