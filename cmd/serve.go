package cmd

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"os"

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

		htr := &view.HTMLTemplateRenderer{}

		server, err := server.NewAppServer(viper.GetString("http_service_address"), csrfKey, viper.GetBool("insecure_dev_mode"), cookieHashKey, cookieBlockKey, dbpool, htr)
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

	serveCmd.Flags().Bool("insecure-dev-mode", false, "Insecure development mode")
	viper.BindPFlag("insecure_dev_mode", serveCmd.Flags().Lookup("insecure-dev-mode"))

	serveCmd.Flags().StringP("database-url", "d", "", "Database URL or DSN")
	viper.BindPFlag("database_url", serveCmd.Flags().Lookup("database-url"))

	serveCmd.Flags().String("html-path", "html", "HTML template path")
	viper.BindPFlag("html_path", serveCmd.Flags().Lookup("html-path"))
}
