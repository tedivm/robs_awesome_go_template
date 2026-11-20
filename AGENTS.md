# AGENTS.md — Rob's Awesome Go Template (Template Maintenance)

**This file covers how to maintain THIS cookiecutter template.** For Go-specific conventions, patterns, and gotchas that apply to the generated projects, see `{{ cookiecutter.project_slug }}/AGENTS.md` and follow those rules when modifying template code.

---

## Critical: Template Syntax Conflicts

Cookiecutter uses Jinja2 (`{{ }}` and `{% %}`). Multiple tools in this template use the SAME delimiters. **Every conflict must be wrapped in `{% raw %}...{% endraw %}` blocks.**

### Conflicting Tools

| Tool | Syntax | Where | Fix |
|------|--------|-------|-----|
| **Go Releaser** | `{{ .Tag }}`, `{{ .Version }}`, `{{ .ProjectName }}` | `.goreleaser.yaml` | Wrap Go Releaser vars in `{% raw %}`; pass cookiecutter values via `env:` block |
| **Cobra CLI** | `{{ .Version }}`, `{{ .Platform }}` | `internal/cli/root.go` (`SetVersionTemplate`) | Wrap in `{% raw %}` |
| **Go html/template** | `{{ .Title }}`, `{{ template "content" . }}` | `static/templates/layout.html` | Wrap in `{% raw %}` |
| **GitHub Actions** | `${{ github.repository }}`, `${{ secrets.X }}` | `.github/workflows/*.yaml` | Wrap in `{% raw %}` (the `$` doesn't prevent Jinja2 parsing) |

### Rule

**Any file containing `{{` that is NOT a `{{ cookiecutter.* }}` reference MUST be wrapped in `{% raw %}` blocks.** Search for `{{ \.` to find conflicts.

---

## Critical: Cookiecutter Hooks

Hooks (`hooks/pre_gen_project.py`, `hooks/post_gen_project.py`) are **processed through Jinja2** before execution.

### Constraints

- **No Python f-strings** with `{}` — Jinja2 tries to parse them. Use string concatenation: `"text " + var`
- **No dict literals** `{}` — same issue. Use `dict()` constructor instead.
- **No set literals** — same issue.
- **Cookiecutter variables** are available as `{{ cookiecutter.key }}` — they get substituted into the Python code before it runs.
- **Environment variables are NOT set** by cookiecutter for hooks. Don't use `os.environ.get("COOKIECUTTER_*")`.

### Pattern for post_gen_project.py

```python
# GOOD — Jinja2 substitutes the value before Python runs
include_web = get_bool("{{ cookiecutter.include_web }}")

# BAD — f-strings confuse Jinja2
print(f"Project: {project_name}")

# GOOD — string concatenation
print("Project: " + project_name)
```

---

## Critical: Conditional File Generation

**Do not use conditional filename suffixes** for files inside the generated project:

```
file.yaml{% if cookiecutter.include_feature %}{% endif %}
```

This creates untracked files with mangled names until cookiecutter processes them, breaks IDE indexing, and makes the template harder to maintain. Use the `post_gen_project.py` hook to remove files when their feature is disabled instead.

**Exception:** Files in the template root (outside `{{ cookiecutter.project_slug }}/`) may use conditional suffixes since they're not part of the generated project. This includes `.github/workflows/` files at the template level.

### Workflow Files Using Conditional Suffixes (Template Root Only)

These files live in the template's own `.github/workflows/` and use conditional suffixes because they're not generated into the project:

| File | Condition |
|------|-----------|
| `.github/workflows/docker.yaml{% if ... %}` | `include_docker` |
| `.github/workflows/integration.yaml{% if ... %}` | `include_orm` |
| `.github/workflows/release.yaml{% if ... %}` | `include_cli or include_web` |

**The `post_gen_project.py` hook handles cleanup** for any files that end up in the generated project but shouldn't be there.

---

## File Modification Checklist

When modifying any template file, verify:

1. **Jinja2 delimiters** — Does the file contain `{{` or `{%` that isn't `{{ cookiecutter.* }}`? If yes, wrap in `{% raw %}`.
2. **Package versions** — Are the versions current? Check with `go list -m -versions <package>`.
3. **Import paths** — Do all imports use `{{ cookiecutter.import_path }}` for local packages?
4. **Conditional markers** — Does the file need a conditional filename suffix?
5. **Hook compatibility** — If modifying hooks, avoid f-strings, dict literals, and `{}` in strings.
6. **Build test** — After changes, run a full smoke test: `cookiecutter --no-input ... && go build ./... && go test ./...`

---

## Smoke Test Command

```bash
cd /tmp && rm -rf go_smoke_test && \
cookiecutter /path/to/template --no-input \
  project_name="Go Smoke Test" \
  project_description="Smoke test" \
  author_name="Test" \
  author_email="test@test.com" \
  github_org="testorg" \
  go_version="1.27" \
  license="MIT" \
  include_web="y" \
  include_orm="y" \
  include_cli="y" \
  include_workers="y" \
  include_docker="y" \
  include_ai="y" && \
cd /tmp/go_smoke_test && \
go build ./... && \
go test ./...
```
