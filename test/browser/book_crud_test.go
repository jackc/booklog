package browser_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/booklog/test/testdata"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgxrecord"
	"github.com/stretchr/testify/require"
)

func TestBookCRUDCycle(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	serverInstance := startServer(t)
	db := serverInstance.DB.Connect(t, ctx)
	page := TestBrowserManager.Acquire(t).Page()

	testdata.CreateUser(t, db, ctx, map[string]any{"username": "john", "password": "mysecret"})
	login(t, ctx, page, serverInstance.Server.URL, "john", "mysecret")

	page.HasContent("a", "New Book")

	page.ClickOn("New Book")
	page.FillIn("Title", "Paradise Lost")
	page.FillIn("Author", "Paradise Lost")
	page.ElementByLabel("Finish Date").MustInputTime(time.Date(2019, 1, 1, 0, 0, 0, 0, time.Local))
	page.ElementByLabel("Format").MustSelect("audio")
	page.ClickOn("Save")

	page.HasContent("dd", "Paradise Lost")

	books, err := pgxrecord.Select(ctx, db, "select * from books", nil, pgx.RowToMap)
	require.NoError(t, err)
	require.Len(t, books, 1)
	require.Equal(t, "Paradise Lost", books[0]["title"])

	page.ClickOn("Edit")
	page.FillIn("Title", "Paradise Regained")
	page.ClickOn("Save")

	page.HasContent("dd", "Paradise Regained")

	page.ClickOn("Delete")
	page.ClickOn("Delete")

	page.HasContent("a", "New Book")
	page.DoesNotHaveContent("a", "Paradise Regained")
}
