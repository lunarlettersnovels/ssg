package generator

import (
	"fmt"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lunarlettersnovels/ssg/internal/db"
)

type Job struct {
	Type          string // "series" or "chapter"
	Series        db.Series
	Chapter       *db.Chapter // Metadata only initially
	Prev          *db.Chapter
	Next          *db.Chapter
	CurrentIndex  int
	TotalChapters int
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
	// Buffer needs to be large enough or we blocking-feed it
	jobs := make(chan Job, 1000)
	var wg sync.WaitGroup

	// Start workers
	workerCount := g.cfg.SSG.Concurrency
	if workerCount <= 0 {
		workerCount = 100
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

	// Dispatcher Goroutine
	// We run this in a goroutine so we can close the channel when done
	// and not block the main thread from waiting on WG if we were using it differently,
	// but here we just block main thread on feeding, then wait.
	// Actually, keeping the feed in main thread is fine as long as workers consume.

	// Iterate Series
	for _, s := range seriesList {
		// 1. Series Page Job
		jobs <- Job{Type: "series", Series: s}

		// 2. Fetch Chapters (Metadata)
		chapters, err := g.repo.GetChaptersBySeriesID(s.ID)
		if err != nil {
			fmt.Printf("Failed to get chapters for series %s: %v\n", s.Slug, err)
			continue
		}

		// 3. Dispatch Chapter Jobs
		total := len(chapters)
		for i := range chapters {
			// We need pointers for Prev/Next
			// Be careful with loop variable scope, but we access by index 'i'
			var prev, next *db.Chapter
			if i > 0 {
				prev = &chapters[i-1]
			}
			if i < total-1 {
				next = &chapters[i+1]
			}

			// Copy chapter to avoid pointer to loop var issues if we used &ch
			current := chapters[i]

			jobs <- Job{
				Type:          "chapter",
				Series:        s,
				Chapter:       &current,
				Prev:          prev,
				Next:          next,
				CurrentIndex:  i + 1,
				TotalChapters: total,
			}
		}
	}

	close(jobs)
	wg.Wait()
	done <- true

	fmt.Printf("\nGenerated %d files.\n", filesGenerated)
	return nil
}

func (g *Generator) processJob(job Job) error {
	if job.Type == "series" {
		return g.renderSeriesPage(job.Series)
	} else if job.Type == "chapter" {
		// Fetch content just-in-time
		fullChapter, err := g.repo.GetChapterContent(job.Chapter.ID)
		if err != nil {
			return err
		}
		if fullChapter == nil {
			return fmt.Errorf("chapter content not found for id %d", job.Chapter.ID)
		}

		return g.renderChapterPage(job.Series, fullChapter, job.Prev, job.Next, job.CurrentIndex, job.TotalChapters)
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
