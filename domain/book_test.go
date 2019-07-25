package domain_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/booklog/domain"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/require"
)

func closeConn(t testing.TB, conn *pgx.Conn) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, conn.Close(ctx))
}

func TestImportBooksFromCSV(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	conn, err := pgx.Connect(ctx, os.Getenv("BOOKLOG_TEST_DB_CONN_STRING"))
	require.NoError(t, err)
	defer closeConn(t, conn)

	tx, err := conn.Begin(ctx, nil)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	var userID int64
	err = tx.QueryRow(ctx, "insert into users(username, password_digest) values('test', 'x') returning id").Scan(&userID)
	require.NoError(t, err)

	in := `Title,Author,Date Finished,Media,
	Paradise Lost ,John Milton ,7/2/2005,,
	The Dilbert Future ,Scott Adams ,7/10/2005,,
	Napoleon The Man Behind the Myth,Adam Zamoyski,6/17/2019,audiobook,`

	err = domain.ImportBooksFromCSV(ctx, tx, userID, strings.NewReader(in))
	require.NoError(t, err)

	var bookCount int64
	err = tx.QueryRow(ctx, "select count(*) from books where user_id=$1", userID).Scan(&bookCount)
	require.NoError(t, err)

	require.EqualValues(t, 3, bookCount)
}

func TestDeleteBookSuccess(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	conn, err := pgx.Connect(ctx, os.Getenv("BOOKLOG_TEST_DB_CONN_STRING"))
	require.NoError(t, err)
	defer closeConn(t, conn)

	tx, err := conn.Begin(ctx, nil)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	var userID int64
	err = tx.QueryRow(ctx, "insert into users(username, password_digest) values('test', 'x') returning id").Scan(&userID)
	require.NoError(t, err)

	var bookID int64
	err = tx.QueryRow(ctx,
		"insert into books(user_id, title, author, finish_date, media) values($1, $2, $3, $4, $5) returning id",
		userID, "Paradise Lost", "John Milton", time.Now(), "book",
	).Scan(&bookID)
	require.NoError(t, err)

	err = domain.DeleteBook(ctx, tx, userID, domain.DeleteBookArgs{ID: bookID})
	require.NoError(t, err)

	var bookCount int64
	err = tx.QueryRow(ctx, "select count(*) from books where user_id=$1", userID).Scan(&bookCount)
	require.NoError(t, err)

	require.EqualValues(t, 0, bookCount)
}

func TestDeleteBookBadFormattedBookIDString(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	conn, err := pgx.Connect(ctx, os.Getenv("BOOKLOG_TEST_DB_CONN_STRING"))
	require.NoError(t, err)
	defer closeConn(t, conn)

	tx, err := conn.Begin(ctx, nil)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	var userID int64
	err = tx.QueryRow(ctx, "insert into users(username, password_digest) values('test', 'x') returning id").Scan(&userID)
	require.NoError(t, err)

	var bookID int64
	err = tx.QueryRow(ctx,
		"insert into books(user_id, title, author, finish_date, media) values($1, $2, $3, $4, $5) returning id",
		userID, "Paradise Lost", "John Milton", time.Now(), "book",
	).Scan(&bookID)
	require.NoError(t, err)

	err = domain.DeleteBook(ctx, tx, userID, domain.DeleteBookArgs{IDString: "123abc"})
	require.Error(t, err)
	require.IsType(t, &domain.NotFoundError{}, err)

	var bookCount int64
	err = tx.QueryRow(ctx, "select count(*) from books where user_id=$1", userID).Scan(&bookCount)
	require.NoError(t, err)

	require.EqualValues(t, 1, bookCount)
}

func TestDeleteBookMissingBookID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	conn, err := pgx.Connect(ctx, os.Getenv("BOOKLOG_TEST_DB_CONN_STRING"))
	require.NoError(t, err)
	defer closeConn(t, conn)

	tx, err := conn.Begin(ctx, nil)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	var userID int64
	err = tx.QueryRow(ctx, "insert into users(username, password_digest) values('test', 'x') returning id").Scan(&userID)
	require.NoError(t, err)

	var bookID int64
	err = tx.QueryRow(ctx,
		"insert into books(user_id, title, author, finish_date, media) values($1, $2, $3, $4, $5) returning id",
		userID, "Paradise Lost", "John Milton", time.Now(), "book",
	).Scan(&bookID)
	require.NoError(t, err)

	err = domain.DeleteBook(ctx, tx, userID, domain.DeleteBookArgs{IDString: "-1"})
	require.Error(t, err)
	require.IsType(t, &domain.NotFoundError{}, err)

	var bookCount int64
	err = tx.QueryRow(ctx, "select count(*) from books where user_id=$1", userID).Scan(&bookCount)
	require.NoError(t, err)

	require.EqualValues(t, 1, bookCount)
}

func TestDeleteBookForbidden(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	conn, err := pgx.Connect(ctx, os.Getenv("BOOKLOG_TEST_DB_CONN_STRING"))
	require.NoError(t, err)
	defer closeConn(t, conn)

	tx, err := conn.Begin(ctx, nil)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	var userID int64
	err = tx.QueryRow(ctx, "insert into users(username, password_digest) values('test', 'x') returning id").Scan(&userID)
	require.NoError(t, err)

	var bookID int64
	err = tx.QueryRow(ctx,
		"insert into books(user_id, title, author, finish_date, media) values($1, $2, $3, $4, $5) returning id",
		userID, "Paradise Lost", "John Milton", time.Now(), "book",
	).Scan(&bookID)
	require.NoError(t, err)

	var otherUserID int64
	err = tx.QueryRow(ctx, "insert into users(username, password_digest) values('otheruser', 'x') returning id").Scan(&otherUserID)
	require.NoError(t, err)

	err = domain.DeleteBook(ctx, tx, otherUserID, domain.DeleteBookArgs{ID: bookID})
	require.Error(t, err)
	require.IsType(t, &domain.ForbiddenError{}, err)

	var bookCount int64
	err = tx.QueryRow(ctx, "select count(*) from books where user_id=$1", userID).Scan(&bookCount)
	require.NoError(t, err)

	require.EqualValues(t, 1, bookCount)
}
