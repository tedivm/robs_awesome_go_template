package renderer_test

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"{{ cookiecutter.import_path }}/internal/renderer"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRenderer(t *testing.T) (*renderer.Renderer, string) {
	t.Helper()

	tmpDir := t.TempDir()

	// Create layout.html
{% raw %}
	layout := `<!DOCTYPE html>
<html>
<head><title>{{ .Title }}</title></head>
<body>{{ template "content" . }}</body>
</html>`
{% endraw %}
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "layout.html"), []byte(layout), 0644))

	// Create page.html
{% raw %}
	page := `{{ define "content" }}<h1>{{ .Heading }}</h1>{{ end }}`
{% endraw %}
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "page.html"), []byte(page), 0644))

	return renderer.New(tmpDir), tmpDir
}

func TestRender(t *testing.T) {
	t.Parallel()

	rendererInst, _ := setupTestRenderer(t)

	w := httptest.NewRecorder()
	rendererInst.Render(w, "page.html", map[string]string{
		"Title":   "Test Page",
		"Heading": "Hello World",
	})

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "Test Page")
	assert.Contains(t, w.Body.String(), "Hello World")
}

func TestRenderPathTraversal(t *testing.T) {
	t.Parallel()

	rendererInst, _ := setupTestRenderer(t)

	w := httptest.NewRecorder()
	rendererInst.Render(w, "../../etc/passwd", nil)

	assert.Equal(t, 400, w.Code)
	assert.Contains(t, w.Body.String(), "invalid template name")
}

func TestRenderMissingTemplate(t *testing.T) {
	t.Parallel()

	rendererInst, _ := setupTestRenderer(t)

	w := httptest.NewRecorder()
	rendererInst.Render(w, "nonexistent.html", map[string]string{"Title": "Test"})

	assert.Equal(t, 500, w.Code)
}

func TestRenderCaching(t *testing.T) {
	t.Parallel()

	rendererInst, _ := setupTestRenderer(t)

	// First render
	w1 := httptest.NewRecorder()
	rendererInst.Render(w1, "page.html", map[string]string{"Title": "Test", "Heading": "First"})
	assert.Equal(t, 200, w1.Code)

	// Second render (should use cached template)
	w2 := httptest.NewRecorder()
	rendererInst.Render(w2, "page.html", map[string]string{"Title": "Test", "Heading": "Second"})
	assert.Equal(t, 200, w2.Code)
	assert.Contains(t, w2.Body.String(), "Second")
}
