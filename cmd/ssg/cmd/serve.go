package cmd

import (
	"fmt"
	"log"
	"net/http"

	"github.com/lunarlettersnovels/ssg/internal/config"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve the generated site",
	Run: func(cmd *cobra.Command, args []string) {
		configFile, _ := cmd.Flags().GetString("config")
		cfg, err := config.Load(configFile)
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		port, _ := cmd.Flags().GetString("port")
		dir := cfg.SSG.OutputDir

		fmt.Printf("Serving %s on http://localhost:%s\n", dir, port)
		log.Fatal(http.ListenAndServe(":"+port, http.FileServer(http.Dir(dir))))
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringP("port", "p", "6969", "Port to serve on")
	serveCmd.Flags().StringP("config", "c", "config.ini", "config file")
}
