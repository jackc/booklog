package browser_test

import (
	"context"
	"testing"

	"github.com/jackc/booklog/test/testdata"
	"github.com/stretchr/testify/require"
)

func TestUserLogsInAndLogsOut(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	serverInstance := startServer(t)
	db := serverInstance.DB.Connect(t, ctx)
	page := TestBrowserManager.Acquire(t).Page()

	testdata.CreateUser(t, db, ctx, map[string]any{"username": "john", "password": "mysecret"})
	login(t, ctx, page, serverInstance.Server.URL, "john", "mysecret")

	page.HasContent("body", "Per Year")

	page.ClickOn("Logout")

	page.HasContent("label", "Username")
	page.HasContent("label", "Password")
}

func TestUserWithInvalidSessionIsLoggedOut(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	serverInstance := startServer(t)
	db := serverInstance.DB.Connect(t, ctx)
	page := TestBrowserManager.Acquire(t).Page()

	testdata.CreateUser(t, db, ctx, map[string]any{"username": "john", "password": "mysecret"})
	login(t, ctx, page, serverInstance.Server.URL, "john", "mysecret")

	page.HasContent("body", "Per Year")

	_, err := db.Exec(ctx, "delete from user_sessions")
	require.NoError(t, err)

	page.ClickOn("New Book")

	page.HasContent("label", "Username")
	page.HasContent("label", "Password")
}
