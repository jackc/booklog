package server

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
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

	tx, err := conn.Begin(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	var userID int64
	err = tx.QueryRow(ctx, "insert into users(username, password_digest) values('test', 'x') returning id").Scan(&userID)
	require.NoError(t, err)

	in := `Title,Author,Date Finished,Format,
	Paradise Lost ,John Milton ,7/2/2005,text,
	The Dilbert Future ,Scott Adams ,7/10/2005,text,
	Napoleon The Man Behind the Myth,Adam Zamoyski,6/17/2019,audio,`

	err = importBooksFromCSV(ctx, tx, userID, strings.NewReader(in))
	require.NoError(t, err)

	var bookCount int64
	err = tx.QueryRow(ctx, "select count(*) from books where user_id=$1", userID).Scan(&bookCount)
	require.NoError(t, err)

	require.EqualValues(t, 3, bookCount)
}
