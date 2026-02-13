package generator

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/lunarlettersnovels/ssg/internal/config"
	"github.com/lunarlettersnovels/ssg/internal/db"
	"github.com/lunarlettersnovels/ssg/internal/templates"
)

type Generator struct {
	cfg      *config.Config
	repo     *db.Repository
	renderer *templates.Renderer
}

func New(cfg *config.Config, repo *db.Repository, renderer *templates.Renderer) *Generator {
	return &Generator{
		cfg:      cfg,
		repo:     repo,
		renderer: renderer,
	}
}

func (g *Generator) Generate() error {
	start := time.Now()
	fmt.Println("Starting generation...")

	// 1. Prepare Output Directory
	if err := g.prepareOutput(); err != nil {
		return err
	}

	// 2. Generate Homepage
	if err := g.generateHomepage(); err != nil {
		return err
	}

	// 3. Generate Series & Chapters (Concurrent)
	if err := g.generateContent(); err != nil {
		return err
	}

	// 4. Generate Sitemap
	// Retrieve series list again or pass it? generateContent fetches it.
	// Let's refactor slightly or just fetch again (cheap).
	seriesList, err := g.repo.GetSeriesList()
	if err == nil {
		g.generateSitemap(seriesList)
	}

	fmt.Printf("Generation completed in %v\n", time.Since(start))
	return nil
}

func (g *Generator) prepareOutput() error {
	// Clean output dir
	if err := os.RemoveAll(g.cfg.SSG.OutputDir); err != nil {
		return fmt.Errorf("failed to clean output dir: %w", err)
	}
	if err := os.MkdirAll(g.cfg.SSG.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output dir: %w", err)
	}

	// Copy Assets
	// Validating that assets exist
	assetsSrc := filepath.Join("public", "assets")
	assetsDest := filepath.Join(g.cfg.SSG.OutputDir, "assets")

	// Create assets dir
	if err := os.MkdirAll(assetsDest, 0755); err != nil {
		return err
	}

	// Copy files
	entries, err := os.ReadDir(assetsSrc)
	if err != nil {
		// It's possible assets aren't built yet, warn but don't fail hard?
		// actually fail hard so user knows.
		return fmt.Errorf("failed to read assets dir (did you build ui?): %w", err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(assetsSrc, entry.Name())
		destPath := filepath.Join(assetsDest, entry.Name())
		if err := copyFile(srcPath, destPath); err != nil {
			return err
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	if _, err := io.Copy(destination, source); err != nil {
		return err
	}
	return nil
}
