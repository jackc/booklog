package cmd

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"os"

	"github.com/jackc/booklog/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start web server",
	Run: func(cmd *cobra.Command, args []string) {
		csrfKey := make([]byte, 32)
		if s := viper.GetString("csrf_key"); len(s) >= 64 {
			h := sha256.Sum256([]byte(s))
			copy(csrfKey, h[:])
		} else {
			fmt.Fprintln(os.Stderr, "CSRF key not set or too short. Using random key.")
			if _, err := io.ReadFull(rand.Reader, csrfKey); err != nil {
				fmt.Fprintf(os.Stderr, "error creating random CSRF key: %v\n", err)
				os.Exit(1)
			}
		}

		server.Serve(viper.GetString("http_service_address"), csrfKey, viper.GetBool("insecure_dev_mode"))
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().StringP("http-service-address", "a", "127.0.0.1:3000", "HTTP service address")
	viper.BindPFlag("http_service_address", serveCmd.Flags().Lookup("http-service-address"))

	serveCmd.Flags().String("csrf-key", "", "CSRF key")
	viper.BindPFlag("csrf_key", serveCmd.Flags().Lookup("csrf-key"))

	serveCmd.Flags().Bool("insecure-dev-mode", false, "Insecure development mode")
	viper.BindPFlag("insecure_dev_mode", serveCmd.Flags().Lookup("insecure-dev-mode"))

	serveCmd.Flags().StringP("database-url", "d", "127.0.0.1:3000", "Database URL or DSN")
	viper.BindPFlag("database_url", serveCmd.Flags().Lookup("database-url"))
}
