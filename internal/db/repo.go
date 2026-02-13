package db

import (
	"database/sql"
	"time"
)

type Series struct {
	ID           uint64
	Slug         string
	Title        string
	ThumbnailURL sql.NullString
	Author       sql.NullString
	Description  sql.NullString
	Status       sql.NullString
	Genre        sql.NullString
	ReleaseYear  sql.NullInt32
	SourceID     uint
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Chapter struct {
	ID            uint64
	SeriesID      uint64
	ChapterNumber float64
	Title         sql.NullString
	Content       sql.NullString
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetSeriesList fetches all series for the homepage/sitemap
func (r *Repository) GetSeriesList() ([]Series, error) {
	query := `SELECT id, slug, title, thumbnail_url, author, description, status, genre, release_year, source_id, created_at, updated_at FROM series ORDER BY updated_at DESC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Series
	for rows.Next() {
		var s Series
		if err := rows.Scan(&s.ID, &s.Slug, &s.Title, &s.ThumbnailURL, &s.Author, &s.Description, &s.Status, &s.Genre, &s.ReleaseYear, &s.SourceID, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, nil
}

// GetSeriesDetail fetches a single series by slug
func (r *Repository) GetSeriesDetail(slug string) (*Series, error) {
	query := `SELECT id, slug, title, thumbnail_url, author, description, status, genre, release_year, source_id, created_at, updated_at FROM series WHERE slug = ?`
	row := r.db.QueryRow(query, slug)

	var s Series
	if err := row.Scan(&s.ID, &s.Slug, &s.Title, &s.ThumbnailURL, &s.Author, &s.Description, &s.Status, &s.Genre, &s.ReleaseYear, &s.SourceID, &s.CreatedAt, &s.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, err
	}
	return &s, nil
}

// GetChaptersBySeriesID fetches all chapters for a series
func (r *Repository) GetChaptersBySeriesID(seriesID uint64) ([]Chapter, error) {
	query := `SELECT id, series_id, chapter_number, title, created_at, updated_at FROM chapters WHERE series_id = ? ORDER BY chapter_number ASC`
	rows, err := r.db.Query(query, seriesID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Chapter
	for rows.Next() {
		var c Chapter
		if err := rows.Scan(&c.ID, &c.SeriesID, &c.ChapterNumber, &c.Title, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		// Content is not fetched here to save memory/bandwidth
		list = append(list, c)
	}
	return list, nil
}

// GetChapterContent fetches a single chapter with content
func (r *Repository) GetChapterContent(id uint64) (*Chapter, error) {
	query := `SELECT id, series_id, chapter_number, title, content, created_at, updated_at FROM chapters WHERE id = ?`
	row := r.db.QueryRow(query, id)

	var c Chapter
	if err := row.Scan(&c.ID, &c.SeriesID, &c.ChapterNumber, &c.Title, &c.Content, &c.CreatedAt, &c.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}
