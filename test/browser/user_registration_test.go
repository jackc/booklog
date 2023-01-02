package browser_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/booklog/test/testdata"
)

func TestUserRegistration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	serverInstance := startServer(t)
	db := serverInstance.DB.Connect(t, ctx)
	page := TestBrowserManager.Acquire(t).Page()

	testdata.CreateUser(t, db, ctx, map[string]any{"username": "john", "password": "mysecret"})

	page.MustNavigate(fmt.Sprintf("%s/user_registration/new", serverInstance.Server.URL))
	page.FillIn("Username", "test")
	page.FillIn("Password", "secret phrase")
	page.ClickOn("Sign up")
	page.HasContent("a", "New Book")
}
