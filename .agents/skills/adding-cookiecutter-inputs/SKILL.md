---
name: adding-cookiecutter-inputs
description: "Use when adding a new cookiecutter input/flag to the template (cookiecutter.json). Covers adding the input, updating template conditionals, post_gen hooks, test configurations, and CI matrix. Don't use for modifying existing inputs or non-cookiecutter template changes."
license: MIT
compatibility: "cookiecutter, Jinja2, Python 3, Go 1.21+"
metadata:
  author: Robert Hafner
  source: "https://github.com/tedivm/robs_awesome_go_template"
---

# Adding Cookiecutter Inputs

## Workflow

1. **Add the input to `cookiecutter.json`:**

   ```json
   "include_newfeature": "y"
   ```

   Use `"y"`/`"n"` for boolean flags (matches existing convention). Set the default to `"n"` for new features until proven stable.

2. **Add conditionals to affected template files:**

   Every file in `{{ cookiecutter.project_slug }}/` that references the new feature needs Jinja2 conditionals:

   ```jinja
   {% if cookiecutter.include_newfeature == "y" %}
   ... feature-specific content ...
   {% endif %}
   ```

   Files to check:
   - `go.mod` — add/remove dependencies
   - `Dockerfile` — add/remove build steps, entrypoints
   - `docker-compose.yaml` — add/remove services
   - `justfile` — add/remove tasks
   - `cmd/*/main.go` — add/remove imports, initialization
   - `internal/*/` — add/remove packages or conditional imports
   - Documentation in `docs/dev/`
   - `.github/workflows/` — add/remove workflow files or job conditionals
   - Any file with `{{ cookiecutter.* }}` that would break without the feature

3. **Handle file/directory cleanup in `hooks/post_gen_project.py`:**

   If the feature has dedicated files/directories that should be removed when disabled:

   ```python
   if not include_newfeature:
       print("[newfeature disabled] Removing newfeature files...")
       remove_path("internal/newfeature")
       remove_path("docs/dev/newfeature.md")
   ```

   If the feature is a prerequisite for another (e.g., CLI requires ORM), add the dependency check:

   ```python
   if not include_cli or not include_orm:
       # remove CLI files
   ```

4. **Create test configurations in `tests/`:**

   New feature combinations MUST be tested. Create YAML files for:

   - The feature alone (if standalone): `tests/newfeature-only.yaml`
   - The feature with its primary partner: `tests/web-newfeature.yaml`
   - The feature with docker (if applicable): `tests/newfeature-docker.yaml`
   - Any edge cases (e.g., feature enabled but its dependencies disabled)

   Each config file needs:
   ```yaml
   default_context:
     project_name: "Newfeature Test"
     # ... all cookiecutter values ...
     include_newfeature: "y"

   required_files:
     - "internal/newfeature/feature.go"

   excluded_files: []
   ```

   **Existing configs may need updating** if the new feature affects them. Review all existing `tests/*.yaml` files for broken required/excluded file lists.

5. **Update the CI workflow matrix in `.github/workflows/test.yaml`:**

   Add the new config names to the matrix:
   ```yaml
   matrix:
     config:
       - bare
       - newfeature-only
       - web-newfeature
       # ... existing configs ...
   ```

6. **Run the full test suite locally:**

   ```bash
   go run tests/test_template.go
   ```

   All configs must pass before committing. If any fail, fix the template conditionals or test configs until clean.

7. **Verify edge cases:**

   Test combinations that exercise the new feature's boundaries:
   - Feature enabled with all other features disabled
   - Feature enabled with each other feature individually
   - Feature enabled with all other features enabled
   - Feature disabled with all other features enabled (verifies cleanup)

## Key constraints

- **New inputs require new tests.** Every added flag expands the combination space. The test matrix must cover the new combinations.
- **Boolean flags use `"y"`/`"n"`** (not `true`/`false`). Check with `== "y"`.
- **Dependencies between features** (e.g., CLI requires ORM) must be enforced in `post_gen_project.py`, not just documented.
- **post_gen cleanup is authoritative.** If a feature is disabled, its files MUST be removed. Don't rely on conditionals alone — conditionals in generated files are for documentation and configs; post_gen handles code directories.
