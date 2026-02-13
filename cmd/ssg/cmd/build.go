package cmd

import (
	"log"

	"github.com/lunarlettersnovels/ssg/internal/config"
	"github.com/lunarlettersnovels/ssg/internal/db"
	"github.com/lunarlettersnovels/ssg/internal/generator"
	"github.com/lunarlettersnovels/ssg/internal/templates"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the static site",
	Long:  `Generates the static site from the database content.`,
	Run: func(cmd *cobra.Command, args []string) {
		configFile, _ := cmd.Flags().GetString("config")

		// Load Config
		cfg, err := config.Load(configFile)
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		// Setup DB
		database, err := db.Connect(cfg.Database.DSN)
		if err != nil {
			log.Fatalf("Failed to connect to DB: %v", err)
		}
		defer database.Close()

		repo := db.NewRepository(database)

		// Setup Templates
		renderer, err := templates.NewRenderer("templates")
		if err != nil {
			log.Fatalf("Failed to load templates: %v", err)
		}

		// Setup Generator
		gen := generator.New(cfg, repo, renderer)

		// Run
		if err := gen.Generate(); err != nil {
			log.Fatalf("Generation failed: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().StringP("config", "c", "config.ini", "config file (default is config.ini)")
}
