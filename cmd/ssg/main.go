package main

import (
	"flag"
	"log"

	"github.com/lunarlettersnovels/ssg/internal/config"
	"github.com/lunarlettersnovels/ssg/internal/db"
	"github.com/lunarlettersnovels/ssg/internal/generator"
	"github.com/lunarlettersnovels/ssg/internal/templates"
)

func main() {
	configPath := flag.String("config", "config.ini", "Path to config file")
	flag.Parse()

	// Load Config
	cfg, err := config.Load(*configPath)
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
}
