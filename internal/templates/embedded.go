package templates

import (
	"embed"
	"fmt"
	"html/template"

	"circles.diy/internal/models"
)

//go:embed html
var templatesFS embed.FS

// parseTemplateFromEmbedded creates a template with layouts, components, and a specific page
func parseTemplateFromEmbedded(name string, pagePath string) (*template.Template, error) {
	// Define custom template functions
	funcMap := template.FuncMap{
		"sub": func(a, b int) int {
			return a - b
		},
		"slice": func(s interface{}, start, end int) interface{} {
			switch v := s.(type) {
			case []models.MediaItem:
				if start < 0 || end > len(v) || start > end {
					return []models.MediaItem{}
				}
				return v[start:end]
			case []interface{}:
				if start < 0 || end > len(v) || start > end {
					return []interface{}{}
				}
				return v[start:end]
			default:
				return s
			}
		},
		"add": func(a, b int) int {
			return a + b
		},
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
	}

	tmpl := template.New(name).Funcs(funcMap)

	// Parse layouts
	layoutFiles, err := templatesFS.ReadDir("html/layouts")
	if err != nil {
		return nil, fmt.Errorf("failed to read layout directory: %v", err)
	}
	for _, file := range layoutFiles {
		if file.IsDir() {
			continue
		}
		content, err := templatesFS.ReadFile("html/layouts/" + file.Name())
		if err != nil {
			return nil, fmt.Errorf("failed to read layout file %s: %v", file.Name(), err)
		}
		_, err = tmpl.New(file.Name()).Parse(string(content))
		if err != nil {
			return nil, fmt.Errorf("failed to parse layout template %s: %v", file.Name(), err)
		}
	}

	// Parse components
	componentFiles, err := templatesFS.ReadDir("html/components")
	if err != nil {
		return nil, fmt.Errorf("failed to read components directory: %v", err)
	}
	for _, file := range componentFiles {
		if file.IsDir() {
			continue
		}
		content, err := templatesFS.ReadFile("html/components/" + file.Name())
		if err != nil {
			return nil, fmt.Errorf("failed to read component file %s: %v", file.Name(), err)
		}
		_, err = tmpl.New(file.Name()).Parse(string(content))
		if err != nil {
			return nil, fmt.Errorf("failed to parse component template %s: %v", file.Name(), err)
		}
	}

	// Parse the specific page
	if pagePath != "" {
		content, err := templatesFS.ReadFile(pagePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read page file %s: %v", pagePath, err)
		}
		_, err = tmpl.New("page").Parse(string(content))
		if err != nil {
			return nil, fmt.Errorf("failed to parse page template %s: %v", pagePath, err)
		}
	}

	return tmpl, nil
}