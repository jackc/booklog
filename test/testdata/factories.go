package testdata

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgxutil"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

var counter atomic.Int64

func CreateUser(t testing.TB, db pgxutil.DB, ctx context.Context, attrs map[string]any) map[string]any {
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

	user, err := pgxutil.InsertRowReturning(ctx, db, pgx.Identifier{"users"}, attrs, "*", pgx.RowToMap)
	require.NoError(t, err)

	return user
}

func CreateBook(t testing.TB, db pgxutil.DB, ctx context.Context, attrs map[string]any) map[string]any {
	user, err := pgxutil.InsertRowReturning(ctx, db, pgx.Identifier{"books"}, attrs, "*", pgx.RowToMap)
	require.NoError(t, err)

	return user
}
