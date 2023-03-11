package cmd

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jackc/booklog/server"
	"github.com/jackc/booklog/view"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start web server",
	Run: func(cmd *cobra.Command, args []string) {
		digestKey := func(size int, keyName string) []byte {
			buf := make([]byte, size)
			if s := viper.GetString(keyName); len(s) >= size {
				h := sha256.Sum256([]byte(s))
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

		csrfKey := digestKey(32, "csrf_key")
		cookieHashKey := digestKey(32, "cookie_hash_key")
		cookieBlockKey := digestKey(32, "cookie_block_key")

		dbpool, err := pgxpool.New(context.Background(), viper.GetString("database_url"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create DB pool: %v\n", err)
			os.Exit(1)
		}

		var devMode = viper.GetBool("dev")
		var reloadHTMLTemplates = viper.GetBool("reload_html_templates")
		var secureCookies = viper.GetBool("secure_cookies")
		if devMode {
			if !viper.IsSet("reload_html_templates") {
				reloadHTMLTemplates = true
			}
			if !viper.IsSet("secure_cookies") {
				secureCookies = false
			}
		}

		frontendPath := viper.GetString("frontend_path")
		var assetMap map[string]string
		if frontendPath != "" {
			var err error
			assetMap, err = view.LoadManifest(filepath.Join(frontendPath, "manifest.json"))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to load manifest.json: %v\n", err)
				os.Exit(1)
			}
		}

		htr := view.NewHTMLTemplateRenderer(viper.GetString("html_template_path"), assetMap, reloadHTMLTemplates)

		server, err := server.NewAppServer(viper.GetString("http_service_address"), csrfKey, secureCookies, cookieHashKey, cookieBlockKey, dbpool, htr, devMode, frontendPath)
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

	serveCmd.Flags().StringP("http-service-address", "a", "127.0.0.1:3000", "HTTP service address")
	viper.BindPFlag("http_service_address", serveCmd.Flags().Lookup("http-service-address"))

	serveCmd.Flags().String("csrf-key", "", "CSRF key")
	viper.BindPFlag("csrf_key", serveCmd.Flags().Lookup("csrf-key"))

	serveCmd.Flags().String("cookie-hash-key", "", "Cookie hash key")
	viper.BindPFlag("cookie_hash_key", serveCmd.Flags().Lookup("cookie-hash-key"))

	serveCmd.Flags().String("cookie-block-key", "", "Cookie block key")
	viper.BindPFlag("cookie_block_key", serveCmd.Flags().Lookup("cookie-block-key"))

	serveCmd.Flags().Bool("secure-cookies", true, "Set Secure flag on cookies")
	viper.BindPFlag("secure_cookies", serveCmd.Flags().Lookup("secure-cookies"))

	serveCmd.Flags().StringP("database-url", "d", "", "Database URL or DSN")
	viper.BindPFlag("database_url", serveCmd.Flags().Lookup("database-url"))

	serveCmd.Flags().String("html-template-path", "html", "HTML template path")
	viper.BindPFlag("html_template_path", serveCmd.Flags().Lookup("html-template-path"))

	serveCmd.Flags().Bool("reload-html-templates", false, "Reload HTML templates")
	viper.BindPFlag("reload_html_templates", serveCmd.Flags().Lookup("reload-html-templates"))

	serveCmd.Flags().Bool("dev", false, "Development mode")
	viper.BindPFlag("dev", serveCmd.Flags().Lookup("dev"))

	serveCmd.Flags().String("frontend-path", "", "Read manifest.json from here and serve ./assets (empty means disable)")
	viper.BindPFlag("frontend_path", serveCmd.Flags().Lookup("frontend-path"))
}
