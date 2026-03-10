package cmd

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/jackc/booklog/server"
	"github.com/jackc/booklog/view"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start web server",
	Run: func(cmd *cobra.Command, args []string) {
		// Helper to get string config with CLI > Env > Default precedence
		getString := func(flagName, envVar string) string {
			flag := cmd.Flags().Lookup(flagName)
			if flag != nil && flag.Changed {
				return flag.Value.String()
			}
			if envValue, ok := os.LookupEnv(envVar); ok {
				return envValue
			}
			if flag != nil {
				return flag.Value.String()
			}
			return ""
		}

		// Helper to get bool config with CLI > Env > Default precedence
		// Returns (value, wasExplicitlySet)
		getBool := func(flagName, envVar string) (bool, bool) {
			flag := cmd.Flags().Lookup(flagName)
			if flag != nil && flag.Changed {
				val, _ := strconv.ParseBool(flag.Value.String())
				return val, true
			}
			if envValue, ok := os.LookupEnv(envVar); ok {
				val, err := strconv.ParseBool(envValue)
				if err == nil {
					return val, true
				}
			}
			if flag != nil {
				val, _ := strconv.ParseBool(flag.Value.String())
				return val, false
			}
			return false, false
		}

		digestKey := func(size int, keyValue, keyName string) []byte {
			buf := make([]byte, size)
			if len(keyValue) >= size {
				h := sha256.Sum256([]byte(keyValue))
				copy(buf, h[:])
			} else {
				fmt.Fprintf(os.Stderr, "%s not set or too short. Using random key.\n", keyName)
				if _, err := io.ReadFull(rand.Reader, buf); err != nil {
					fmt.Fprintf(os.Stderr, "error creating random %s: %v\n", keyName, err)
					os.Exit(1)
				}
			}
			return buf
		}

		csrfKey := digestKey(32, getString("csrf-key", "CSRF_KEY"), "csrf_key")
		cookieHashKey := digestKey(32, getString("cookie-hash-key", "COOKIE_HASH_KEY"), "cookie_hash_key")
		cookieBlockKey := digestKey(32, getString("cookie-block-key", "COOKIE_BLOCK_KEY"), "cookie_block_key")

		dbpool, err := pgxpool.New(context.Background(), getString("database-url", "DATABASE_URL"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create DB pool: %v\n", err)
			os.Exit(1)
		}

		devMode, _ := getBool("dev", "DEV")
		reloadHTMLTemplates, reloadHTMLTemplatesSet := getBool("reload-html-templates", "RELOAD_HTML_TEMPLATES")
		secureCookies, secureCookiesSet := getBool("secure-cookies", "SECURE_COOKIES")

		if devMode {
			if !reloadHTMLTemplatesSet {
				reloadHTMLTemplates = true
			}
			if !secureCookiesSet {
				secureCookies = false
			}
		}

		frontendPath := getString("frontend-path", "FRONTEND_PATH")
		var assetMap map[string]string
		if frontendPath != "" {
			var err error
			assetMap, err = view.LoadManifest(filepath.Join(frontendPath, "manifest.json"))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to load manifest.json: %v\n", err)
				os.Exit(1)
			}
		}

		htr := view.NewHTMLTemplateRenderer(getString("html-template-path", "HTML_TEMPLATE_PATH"), assetMap, reloadHTMLTemplates)

		server, err := server.NewAppServer(
			getString("http-service-address", "HTTP_SERVICE_ADDRESS"),
			csrfKey,
			secureCookies,
			cookieHashKey,
			cookieBlockKey,
			dbpool,
			htr,
			devMode,
			frontendPath,
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not create web server: %v\n", err)
			os.Exit(1)
		}

		err = server.Serve()
		if err != nil {
			os.Stderr.WriteString("Could not start web server!\n")
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().StringP("http-service-address", "a", "127.0.0.1:3000", "HTTP service address (env: HTTP_SERVICE_ADDRESS)")
	serveCmd.Flags().String("csrf-key", "", "CSRF key (env: CSRF_KEY)")
	serveCmd.Flags().String("cookie-hash-key", "", "Cookie hash key (env: COOKIE_HASH_KEY)")
	serveCmd.Flags().String("cookie-block-key", "", "Cookie block key (env: COOKIE_BLOCK_KEY)")
	serveCmd.Flags().Bool("secure-cookies", true, "Set Secure flag on cookies (env: SECURE_COOKIES)")
	serveCmd.Flags().StringP("database-url", "d", "", "Database URL or DSN (env: DATABASE_URL)")
	serveCmd.Flags().String("html-template-path", "html", "HTML template path (env: HTML_TEMPLATE_PATH)")
	serveCmd.Flags().Bool("reload-html-templates", false, "Reload HTML templates (env: RELOAD_HTML_TEMPLATES)")
	serveCmd.Flags().Bool("dev", false, "Development mode (env: DEV)")
	serveCmd.Flags().String("frontend-path", "", "Read manifest.json from here and serve ./assets (env: FRONTEND_PATH)")
}
