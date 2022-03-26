package data_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/booklog/data"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func closeConn(t testing.TB, conn *pgx.Conn) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, conn.Close(ctx))
}

func TestDeleteBookSuccess(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	conn, err := pgx.Connect(ctx, os.Getenv("BOOKLOG_TEST_DB_CONN_STRING"))
	require.NoError(t, err)
	defer closeConn(t, conn)

	tx, err := conn.Begin(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	var userID int64
	err = tx.QueryRow(ctx, "insert into users(username, password_digest) values('test', 'x') returning id").Scan(&userID)
	require.NoError(t, err)

	var bookID int64
	err = tx.QueryRow(ctx,
		"insert into books(user_id, title, author, finish_date, format) values($1, $2, $3, $4, $5) returning id",
		userID, "Paradise Lost", "John Milton", time.Now(), "book",
	).Scan(&bookID)
	require.NoError(t, err)

	err = data.DeleteBook(ctx, tx, bookID)
	require.NoError(t, err)

	var bookCount int64
	err = tx.QueryRow(ctx, "select count(*) from books where user_id=$1", userID).Scan(&bookCount)
	require.NoError(t, err)

	require.EqualValues(t, 0, bookCount)
}

func TestDeleteBookMissingBookID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	conn, err := pgx.Connect(ctx, os.Getenv("BOOKLOG_TEST_DB_CONN_STRING"))
	require.NoError(t, err)
	defer closeConn(t, conn)

	tx, err := conn.Begin(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	var userID int64
	err = tx.QueryRow(ctx, "insert into users(username, password_digest) values('test', 'x') returning id").Scan(&userID)
	require.NoError(t, err)

	var bookID int64
	err = tx.QueryRow(ctx,
		"insert into books(user_id, title, author, finish_date, format) values($1, $2, $3, $4, $5) returning id",
		userID, "Paradise Lost", "John Milton", time.Now(), "book",
	).Scan(&bookID)
	require.NoError(t, err)

	err = data.DeleteBook(ctx, tx, -1)
	require.Error(t, err)
	require.IsType(t, &data.NotFoundError{}, err)

	var bookCount int64
	err = tx.QueryRow(ctx, "select count(*) from books where user_id=$1", userID).Scan(&bookCount)
	require.NoError(t, err)

	require.EqualValues(t, 1, bookCount)
}
