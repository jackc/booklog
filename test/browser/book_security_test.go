package browser_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/booklog/test/testdata"
	"github.com/stretchr/testify/require"
)

func TestBooksIndexRedirectsAnonymouseUsersToLogin(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	serverInstance := startServer(t)
	db := serverInstance.DB.Connect(t, ctx)
	page := TestBrowserManager.Acquire(t).Page()

	user := testdata.CreateUser(t, db, ctx, map[string]any{"username": "test", "password": "secret phrase"})
	testdata.CreateBook(t, db, ctx, map[string]any{"user_id": user["id"], "title": "Foo", "author": "Bar", "finish_date": time.Date(2019, 1, 1, 0, 0, 0, 0, time.Local), "format": "text"})
	page.MustNavigate(fmt.Sprintf("%s/users/test/books", serverInstance.Server.URL))
	page.HasContent("form", "Login")
	require.Equal(t, fmt.Sprintf("%s/login", serverInstance.Server.URL), page.MustInfo().URL)
}

func TestBooksIndexIsForbiddenToOtherUsers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	serverInstance := startServer(t)
	db := serverInstance.DB.Connect(t, ctx)
	page := TestBrowserManager.Acquire(t).Page()

	user := testdata.CreateUser(t, db, ctx, map[string]any{"username": "test", "password": "secret phrase"})
	testdata.CreateBook(t, db, ctx, map[string]any{"user_id": user["id"], "title": "Foo", "author": "Bar", "finish_date": time.Date(2019, 1, 1, 0, 0, 0, 0, time.Local), "format": "text"})

	testdata.CreateUser(t, db, ctx, map[string]any{"username": "other", "password": "secret phrase"})
	login(t, ctx, page, serverInstance.Server.URL, "other", "secret phrase")
	page.MustNavigate(fmt.Sprintf("%s/users/test/books", serverInstance.Server.URL))
	page.HasContent("body", "Forbidden")
}
