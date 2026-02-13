package templates

import (
	"fmt"
	"html/template"
	"io"
	"regexp"
	"strings"
)

type Renderer struct {
	home    *template.Template
	series  *template.Template
	chapter *template.Template
}

func NewRenderer(templateDir string) (*Renderer, error) {
	funcs := template.FuncMap{
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"grad": func(id uint64) string {
			hue := float64(id * 137) // Simplified math for int
			hue = float64(int(hue) % 360)
			return fmt.Sprintf("linear-gradient(135deg, hsl(%.0f, 40%%, 80%%) 0%%, hsl(%.0f, 45%%, 70%%) 100%%)", hue, hue)
		},
		"abbr": func(s string) string {
			if len(s) >= 2 {
				return strings.ToUpper(s[:2])
			}
			return strings.ToUpper(s)
		},
		"split": strings.Split,
		"stripImages": func(s string) string {
			re := regexp.MustCompile(`<img[^>]*>`)
			return re.ReplaceAllString(s, "")
		},
		"year": func() int {
			return 2026 // Hardcoded as per prompt context or use time.Now()
		},
	}

	// Helper to parse base layout + specific page
	parse := func(page string) (*template.Template, error) {
		// New("layout.html") matters if layout.html is the base
		t := template.New("layout.html").Funcs(funcs)
		return t.ParseFiles(
			templateDir+"/layout.html",
			templateDir+"/"+page,
		)
	}

	// 1. Home
	home, err := parse("index.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse home template: %w", err)
	}

	// 2. Series
	series, err := parse("series.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse series template: %w", err)
	}

	// 3. Chapter (Standalone)
	chapter, err := template.New("chapter.html").Funcs(funcs).ParseFiles(
		templateDir + "/chapter.html",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse chapter template: %w", err)
	}

	return &Renderer{home: home, series: series, chapter: chapter}, nil
}

func (r *Renderer) Render(w io.Writer, name string, data interface{}) error {
	switch name {
	case "index.html":
		return r.home.ExecuteTemplate(w, "layout.html", data)
	case "series.html":
		return r.series.ExecuteTemplate(w, "layout.html", data)
	case "chapter.html":
		return r.chapter.ExecuteTemplate(w, "chapter.html", data)
	default:
		return fmt.Errorf("unknown template: %s", name)
	}
}
