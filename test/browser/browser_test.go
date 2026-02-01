package browser_test

import (
	"context"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/go-rod/rod"
	"github.com/jackc/booklog/server"
	"github.com/jackc/booklog/test/testbrowser"
	"github.com/jackc/booklog/test/testutil"
	"github.com/jackc/booklog/view"
	"github.com/jackc/testdb"
	"github.com/stretchr/testify/require"
)

var concurrentChan chan struct{}
var TestDBManager *testdb.Manager
var baseBrowser *rod.Browser
var TestBrowserManager *testbrowser.Manager

func TestMain(m *testing.M) {
	maxConcurrent := 1
	if n, err := strconv.ParseInt(os.Getenv("MAX_CONCURRENT_BROWSER_TESTS"), 10, 32); err == nil {
		maxConcurrent = int(n)
	}
	if maxConcurrent < 1 {
		fmt.Println("MAX_CONCURRENT_BROWSER_TESTS must be greater than 0")
		os.Exit(1)
	}
	concurrentChan = make(chan struct{}, maxConcurrent)

	TestDBManager = testutil.InitTestDBManager(m)

	var err error
	TestBrowserManager, err = testbrowser.NewManager(testbrowser.ManagerConfig{})
	if err != nil {
		fmt.Println("Failed to initialize TestBrowserManager")
		os.Exit(1)
	}

	os.Exit(m.Run())
}

type serverInstanceT struct {
	Server *httptest.Server
	DB     *testdb.DB
}

func startServer(t *testing.T) *serverInstanceT {
	ctx := context.Background()
	db := TestDBManager.AcquireDB(t, ctx)

	csrfKey := make([]byte, 32)
	cookieHashKey := make([]byte, 32)
	cookieBlockKey := make([]byte, 32)

	assetMap, err := view.LoadManifest(filepath.Join("..", "..", "build", "frontend", "manifest.json"))
	require.NoError(t, err)

	handler, err := server.NewAppServer("127.0.0.1:0", csrfKey, false, cookieHashKey, cookieBlockKey, db.PoolConnect(t, ctx), view.NewHTMLTemplateRenderer("../../html", assetMap, false), false, "../../build/frontend")
	require.NoError(t, err)

	server := httptest.NewServer(handler)
	t.Cleanup(func() {
		server.Close()
	})

	instance := &serverInstanceT{
		Server: server,
		DB:     db,
	}

	return instance
}

func login(t *testing.T, ctx context.Context, page *testbrowser.Page, appHost, username, password string) {
	page.MustNavigate(fmt.Sprintf("%s/login", appHost))

	page.Within("form", func(scope *testbrowser.Scope) {
		scope.FillIn("Username", username)
		scope.FillIn("Password", password)
		scope.ClickOn("Login")
	})

	page.HasContent("body", "New Book")
}
