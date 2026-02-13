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
	Type           string
	Series         db.Series
	Chapter        *db.Chapter
	Prev           *db.Chapter
	Next           *db.Chapter
	CurrentIndex   int
	TotalChapters  int
	SeriesChapters []db.Chapter
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
	seriesList, err := g.repo.GetSeriesList()
	if err != nil {
		return err
	}

	fmt.Println("Fetching all chapters from DB...")
	allChaptersMap, err := g.repo.GetAllChaptersGrouped()
	if err != nil {
		return fmt.Errorf("failed to bulk fetch chapters: %w", err)
	}
	fmt.Printf("Loaded chapters for %d series. Starting generation...\n", len(allChaptersMap))

	// Create job channel
	jobs := make(chan Job, 1000)
	var wg sync.WaitGroup

	// Start workers
	workerCount := g.cfg.SSG.Concurrency
	if workerCount <= 0 {
		workerCount = 100
	}

	var filesGenerated uint64

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

	// Dispatch Jobs
	for _, s := range seriesList {
		chapters := allChaptersMap[s.ID]
		if chapters == nil {
			chapters = []db.Chapter{}
		}

		jobs <- Job{Type: "series", Series: s, SeriesChapters: chapters}

		total := len(chapters)
		for i := range chapters {
			var prev, next *db.Chapter
			if i > 0 {
				prev = &chapters[i-1]
			}
			if i < total-1 {
				next = &chapters[i+1]
			}

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
		return g.renderSeriesPage(job.Series, job.SeriesChapters)
	} else if job.Type == "chapter" {
		return g.renderChapterPage(job.Series, job.Chapter, job.Prev, job.Next, job.CurrentIndex, job.TotalChapters)
	}
	return nil
}

func (g *Generator) renderSeriesPage(series db.Series, chapters []db.Chapter) error {
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

	path := filepath.Join("novel", series.Slug, "chapter", fmt.Sprintf("%d.html", chapter.ID))
	return g.renderToFile(path, "chapter.html", data)
}
