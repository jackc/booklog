package testdata

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgxutil"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

var counter atomic.Int64

type DB interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func CreateUser(t testing.TB, db DB, ctx context.Context, attrs map[string]any) map[string]any {
	var password []byte
	if s, ok := attrs["password"].(string); ok {
		password = []byte(s)
		delete(attrs, "password")
	} else {
		password = []byte("password")
	}
	pwDigest, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	require.NoError(t, err)
	attrs["password_digest"] = pwDigest

	if _, ok := attrs["username"]; !ok {
		attrs["username"] = "test"
	}

	user, err := pgxutil.Insert(ctx, db, "users", attrs)
	require.NoError(t, err)

	return user
}
