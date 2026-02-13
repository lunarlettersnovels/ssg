package generator

import (
	"os"
	"path/filepath"
)

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
