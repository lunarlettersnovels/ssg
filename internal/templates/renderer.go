package templates

import (
	"fmt"
	"html/template"
	"io"
	"regexp"
	"strings"
)

type Renderer struct {
	layout  *template.Template
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

	// Parse Layout + Inner templates (Home, Series)
	layout, err := template.New("layout.html").Funcs(funcs).ParseFiles(
		templateDir+"/layout.html",
		templateDir+"/index.html",
		templateDir+"/series.html",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse layout templates: %w", err)
	}

	// Parse Standalone Chapter template
	chapter, err := template.New("chapter.html").Funcs(funcs).ParseFiles(
		templateDir + "/chapter.html",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse chapter template: %w", err)
	}

	return &Renderer{layout: layout, chapter: chapter}, nil
}

func (r *Renderer) Render(w io.Writer, name string, data interface{}) error {
	if name == "chapter.html" {
		return r.chapter.Execute(w, data)
	}
	return r.layout.ExecuteTemplate(w, name, data)
}
