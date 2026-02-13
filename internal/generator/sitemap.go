package generator

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lunarlettersnovels/ssg/internal/db"
)

func (g *Generator) generateSitemap(seriesList []db.Series) error {
	f, err := os.Create(filepath.Join(g.cfg.SSG.OutputDir, "sitemap.xml"))
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintln(f, `<?xml version="1.0" encoding="UTF-8"?>`)
	fmt.Fprintln(f, `<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`)

	fmt.Fprintf(f, "  <url>\n    <loc>%s/</loc>\n    <changefreq>daily</changefreq>\n    <priority>1.0</priority>\n  </url>\n", g.cfg.SSG.BaseURL)

	for _, s := range seriesList {
		fmt.Fprintf(f, "  <url>\n    <loc>%s/novel/%s</loc>\n    <lastmod>%s</lastmod>\n    <changefreq>daily</changefreq>\n    <priority>0.8</priority>\n  </url>\n",
			g.cfg.SSG.BaseURL, s.Slug, s.UpdatedAt.Format("2006-01-02"))
	}

	fmt.Fprintln(f, `</urlset>`)
	return nil
}
