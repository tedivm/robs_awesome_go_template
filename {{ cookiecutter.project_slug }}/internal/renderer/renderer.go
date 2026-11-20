package renderer

import (
	"errors"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// funcMap provides common template functions.
var funcMap = template.FuncMap{
	"truncate": func(s string, max int) string {
		if len(s) <= max {
			return s
		}
		return s[:max] + "..."
	},
	"date": func(t time.Time, layout string) string {
		if layout == "" {
			layout = time.RFC3339
		}
		return t.Format(layout)
	},
	"upper": strings.ToUpper,
	"lower": strings.ToLower,
}

// Renderer handles HTML template rendering with layout inheritance.
type Renderer struct {
	mu        sync.Mutex
	templates map[string]*template.Template
	dir       string
	layout    string // layout template name (default: "layout.html")
}

// New creates a new template renderer.
func New(dir string) *Renderer {
	return &Renderer{
		templates: make(map[string]*template.Template),
		dir:       dir,
		layout:    "layout.html",
	}
}

// WithLayout sets the layout template name (default: "layout.html").
func (r *Renderer) WithLayout(name string) *Renderer {
	r.layout = name
	return r
}

// Render executes a template and writes to the response writer.
func (r *Renderer) Render(w http.ResponseWriter, name string, data any) {
	// Prevent path traversal
	if strings.Contains(name, "..") {
		http.Error(w, "invalid template name", http.StatusBadRequest)
		return
	}

	tmpl, err := r.getTemplate(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Execute the layout template (it includes the page via the content block)
	if err := tmpl.ExecuteTemplate(w, r.layout, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// getTemplate loads and parses templates with caching.
func (r *Renderer) getTemplate(name string) (*template.Template, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if tmpl, ok := r.templates[name]; ok {
		return tmpl, nil
	}

	// Validate the resolved path stays within the template directory
	page := filepath.Clean(name)
	layoutPath := filepath.Join(r.dir, r.layout)
	pagePath := filepath.Join(r.dir, page)

	// Ensure both paths are within the template directory
	realDir, err := filepath.EvalSymlinks(r.dir)
	if err != nil {
		realDir = r.dir
	}
	realPage, err := filepath.EvalSymlinks(pagePath)
	if err == nil {
		if !strings.HasPrefix(realPage, realDir) {
			return nil, errors.New("template path escapes directory")
		}
	}

	// ParseFiles creates named templates for each file.
	// We create a root template with FuncMap and parse both files into it.
	tmpl, err := template.New("root").Funcs(funcMap).ParseFiles(layoutPath, pagePath)
	if err != nil {
		return nil, err
	}

	r.templates[name] = tmpl
	return tmpl, nil
}
