package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lunarlettersnovels/ssg/internal/db"
)

type Job struct {
	Type     string // "series" or "chapter"
	Series   db.Series
	Chapter  *db.Chapter // Nil if series job
	SeriesID uint64      // Used for chapter job
}

func (g *Generator) generateHomepage() error {
	seriesList, err := g.repo.GetSeriesList()
	if err != nil {
		return fmt.Errorf("failed to fetch series list: %w", err)
	}

	data := struct {
		Series []db.Series
	}{
		Series: seriesList,
	}

	return g.renderToFile("index.html", "index.html", data)
}

func (g *Generator) generateContent() error {
	// Fetch all series
	seriesList, err := g.repo.GetSeriesList()
	if err != nil {
		return err
	}

	// Create job channel
	jobs := make(chan Job, len(seriesList)*100) // Buffer estimate
	var wg sync.WaitGroup

	// Start workers
	workerCount := g.cfg.SSG.Concurrency
	if workerCount <= 0 {
		workerCount = 10
	}

	// Stats
	var filesGenerated uint64

	// Progress Monitor
	done := make(chan bool)
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				count := atomic.LoadUint64(&filesGenerated)
				fmt.Printf("\rGenerated %d files...", count)
			case <-done:
				return
			}
		}
	}()

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				if err := g.processJob(job); err != nil {
					fmt.Printf("\nError processing job %s: %v\n", job.Type, err)
				} else {
					atomic.AddUint64(&filesGenerated, 1)
				}
			}
		}()
	}

	// Enqueue Series Jobs
	// We need to fetch chapters for each series to enqueue chapter jobs too.
	// Optimize: Generate series page, AND enqueue chapter jobs.

	for _, s := range seriesList {
		// Enqueue Series Page Generation
		jobs <- Job{Type: "series", Series: s}

		// Fetch chapters for this series
		// Note: This might be heavy to do in main thread.
		// Better: Worker does FetchChapters and enqueues chapter jobs?
		// Or: We do it here. Let's do it here for simplicity of flow,
		// but for 10k pages + concurrency as requested, let's distribute.
	}

	// Wait? No, we need to close channel.
	// Actually, if we want workers to discover chapters, we need a separate dispatch mechanism or waitgroup usage pattern.

	// Simple approach:
	// Just feed Series jobs.
	// The Series Worker will:
	// 1. Generate Series Page
	// 2. Fetch Chapters
	// 3. Generate Chapter Pages (Here directly or enqueue?)
	// Direct generation in worker seems fine if we have enough workers.
	// But "Generate tens of thousands of pages in an instant" implies high concurrency.
	// Let's make "SeriesJob" also spawn "ChapterJobs" if we want granularity,
	// but keeping it simple: A worker handles a whole series (Series Page + All Chapters).
	// This reduces DB contention on "GetChapters" since we do it once per series.

	close(jobs)
	wg.Wait()
	done <- true

	fmt.Printf("Generated %d files.\n", filesGenerated)
	return nil
}

func (g *Generator) processJob(job Job) error {
	if job.Type == "series" {
		// 1. Generate Series Page
		if err := g.renderSeriesPage(job.Series); err != nil {
			return err
		}

		// 2. Fetch Chapters
		chapters, err := g.repo.GetChaptersBySeriesID(job.Series.ID)
		if err != nil {
			return fmt.Errorf("failed to get chapters for series %s: %w", job.Series.Slug, err)
		}

		// 3. Generate Chapter Pages
		// Loop through chapters and generate.
		for i, ch := range chapters {
			// Fetch full content (Repo method GetChapterContent)
			// Optimization: GetChaptersBySeriesID only got metadata.
			// We need content now.
			fullChapter, err := g.repo.GetChapterContent(ch.ID)
			if err != nil {
				return err
			}
			if fullChapter == nil {
				continue
			}

			// Prev/Next logic
			var prev, next *db.Chapter
			if i > 0 {
				prev = &chapters[i-1]
			}
			if i < len(chapters)-1 {
				next = &chapters[i+1]
			}

			if err := g.renderChapterPage(job.Series, fullChapter, prev, next, i+1, len(chapters)); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *Generator) renderSeriesPage(series db.Series) error {
	chapters, err := g.repo.GetChaptersBySeriesID(series.ID)
	if err != nil {
		return err
	}

	data := struct {
		Series   db.Series
		Chapters []db.Chapter
	}{
		Series:   series,
		Chapters: chapters,
	}

	path := filepath.Join("novel", series.Slug, "index.html")
	return g.renderToFile(path, "series.html", data)
}

func (g *Generator) renderChapterPage(series db.Series, chapter *db.Chapter, prev, next *db.Chapter, index, total int) error {
	data := struct {
		Series        db.Series
		Chapter       *db.Chapter
		PrevChapter   *db.Chapter
		NextChapter   *db.Chapter
		CurrentIndex  int
		TotalChapters int
	}{
		Series:        series,
		Chapter:       chapter,
		PrevChapter:   prev,
		NextChapter:   next,
		CurrentIndex:  index,
		TotalChapters: total,
	}

	path := filepath.Join("novel", series.Slug, "chapter", fmt.Sprintf("%d", chapter.ID), "index.html")
	return g.renderToFile(path, "chapter.html", data)
}

func (g *Generator) renderToFile(relPath, templateName string, data interface{}) error {
	fullPath := filepath.Join(g.cfg.SSG.OutputDir, relPath)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	f, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return g.renderer.Render(f, templateName, data)
}
