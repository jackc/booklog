package cmd

import (
	"github.com/spf13/viper"

	"github.com/jackc/booklog/server"
	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start web server",
	Run: func(cmd *cobra.Command, args []string) {
		server.Serve(viper.GetString("http_service_address"))
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	rootCmd.PersistentFlags().StringP("http-service-address", "a", "127.0.0.1:3000", "HTTP service address")
	viper.BindPFlag("http_service_address", rootCmd.PersistentFlags().Lookup("http-service-address"))
}
