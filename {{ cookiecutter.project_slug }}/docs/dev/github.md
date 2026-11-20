# GitHub — CI/CD Workflows

{{ cookiecutter.project_name }} uses GitHub Actions for continuous integration and deployment. Each check runs as an independent workflow for fast feedback and isolated failures.

## Workflows

### Lint

| Property | Value |
|----------|-------|
| File | `.github/workflows/lint.yaml` |
| Triggers | Push to `main`, pull requests to `main` |
| Tool | golangci-lint |
| Duration | ~1-2 minutes |

Runs golangci-lint against all packages. Checks for code style, potential bugs, unused code, and static analysis issues.

**Interpreting failures:**

- **Error on a specific line:** Fix the reported issue. The error message includes the linter name and suggestion.
- **"too many issues":** Run `just lint` locally to see all issues, then fix them incrementally.
- **New linter added:** Update `.golangci.yml` to disable or configure the new linter if needed.

### Test

| Property | Value |
|----------|-------|
| File | `.github/workflows/test.yaml` |
| Triggers | Push to `main`, pull requests to `main` |
| Tool | `go test` with race detector |
| Duration | ~2-3 minutes |

Runs all tests with `-race` to detect data races. Uploads coverage to Codecov.

**Interpreting failures:**

- **Test failure:** The output shows the failing test name and assertion. Run `just test-verbose` locally to reproduce.
- **Race detected:** The race detector reports concurrent access to shared memory. Fix the data race — don't disable the detector.
- **Coverage threshold:** If coverage drops below the configured threshold, add tests or adjust the threshold.

### Build

| Property | Value |
|----------|-------|
| File | `.github/workflows/build.yaml` |
| Triggers | Push to `main`, pull requests to `main` |
| Tool | `go build` |
| Duration | ~1 minute |

Verifies the project compiles. Catches import errors, type mismatches, and missing dependencies before other workflows run.

**Interpreting failures:**

- **Compilation error:** Fix the reported error. This is the fastest feedback — if build fails, lint and test won't run.
- **Missing dependency:** Run `go mod tidy` and commit the updated `go.mod` and `go.sum`.{% if cookiecutter.include_docker == "y" %}

### Docker

| Property | Value |
|----------|-------|
| File | `.github/workflows/docker.yaml` |
| Triggers | Push to `main`, tags matching `v*`, pull requests to `main` |
| Tool | Docker Buildx |
| Duration | ~5-10 minutes (multi-arch) |

Builds Docker images for `linux/amd64` and `linux/arm64`. Pushes to GitHub Container Registry on pushes to main and tags. PRs build without pushing.

**Interpreting failures:**

- **Build error:** The Dockerfile has an issue. Run `just docker-build` locally to reproduce.
- **Push permission denied:** Ensure the workflow has `packages: write` permission and runs on a push or tag event.
- **QEMU timeout:** Multi-arch builds via emulation can time out. Check if the build is stuck on the arm64 stage.{% endif %}
### Security

| Property | Value |
|----------|-------|
| File | `.github/workflows/security.yaml` |
| Triggers | Push to `main`, pull requests to `main`, weekly (Monday 6 AM UTC) |
| Tools | govulncheck, Trivy |
| Duration | ~3-5 minutes |

Two checks:

1. **govulncheck:** Scans Go code for known vulnerabilities in dependencies. Only reports vulnerabilities your code actually reaches (not just transitive dependencies).{% if cookiecutter.include_docker == "y" %}
2. **Trivy:** Scans the Docker image for OS-level vulnerabilities, misconfigurations, and dependency issues.{% endif %}
**Interpreting failures:**

- **govulncheck failure:** Update the vulnerable dependency. Dependabot may have already opened a PR.{% if cookiecutter.include_docker == "y" %}
- **Trivy failure:** Update the base image or fix the misconfiguration. Check the severity — HIGH and CRITICAL fail the workflow.{% endif %}
- **Weekly scheduled run fails:** Review even if no PR triggered it. Dependencies may have new vulnerabilities.{% if cookiecutter.include_web == "y" or cookiecutter.include_cli == "y" %}

### Release

| Property | Value |
|----------|-------|
| File | `.github/workflows/release.yaml` |
| Triggers | Tags matching `v*` |
| Tool | Go Releaser, Syft |
| Duration | ~10-15 minutes |

Produces release artifacts and publishes a GitHub Release.

## Release Process

### Creating a Release

```bash
# Create an annotated tag
git tag -a v1.0.0 -m "Release v1.0.0"

# Push the tag
git push --tags
```

The release workflow triggers automatically and produces:

| Artifact | Description |
|----------|-------------|
{%- if cookiecutter.include_web == "y" %}
| Binaries | `{{ cookiecutter.project_slug }}_linux_amd64`, `{{ cookiecutter.project_slug }}_linux_arm64`, `{{ cookiecutter.project_slug }}_darwin_amd64`, `{{ cookiecutter.project_slug }}_darwin_arm64`, `{{ cookiecutter.project_slug }}_windows_amd64.exe` |
{%- endif %}{%- if cookiecutter.include_cli == "y" %}
| Binaries | `cli_linux_amd64`, `cli_linux_arm64`, `cli_darwin_amd64`, `cli_darwin_arm64`, `cli_windows_amd64.exe` |
{%- endif %}
| Archives | `.tar.gz` for Linux/macOS, `.zip` for Windows |
| Checksums | `SHA256SUMS` file |
| SBOM | `sbom.spdx.json` (CycloneDX format) |
| GitHub Release | All artifacts attached, changelog from commits |

### Version Injection

The binary version is set from the git tag via `-ldflags`.{% if cookiecutter.include_web == "y" %} The `{{ cookiecutter.project_slug }} version` command prints the version, commit, and build date.{% endif %}{% if cookiecutter.include_cli == "y" %} The `cli version` command prints the version, commit, and build date.{% endif %}{% if cookiecutter.include_docker == "y" %}

### Cosign Signing

The release workflow signs container images with cosign using keyless signing via Fulcio CA and Rekor transparency log. Verify signatures:

```bash
cosign verify ghcr.io/{{ cookiecutter.github_org }}/{{ cookiecutter.project_slug }}:v1.0.0
```{% endif %}{% endif %}
## Pre-Commit Hooks

The project uses [pre-commit](https://pre-commit.com/) to run checks before each commit:

```yaml
repos:
  - repo: local
    hooks:
      - id: golangci-lint
        name: golangci-lint
        entry: just lint-check
        language: system
        types: [go]
        pass_filenames: false
      - id: go-fmt
        name: go fmt
        entry: just fmt-check
        language: system
        types: [go]
        pass_filenames: false
      - id: go-vet
        name: go vet
        entry: just vet
        language: system
        types: [go]
        pass_filenames: false
```

### Installing Pre-Commit

```bash
# Install pre-commit
pip install pre-commit

# Install hooks
pre-commit install
```

Hooks run automatically on `git commit`. They check only changed Go files for speed.

### Bypassing Pre-Commit

```bash
# Skip hooks for a commit (use sparingly)
git commit --no-verify
```

## Dependabot

The project uses Dependabot for automated dependency updates:

```yaml
version: 2

updates:
  - package-ecosystem: "github-actions"
    directory: "/"
    schedule:
      interval: "weekly"

  - package-ecosystem: "gomod"
    directory: "/"
    schedule:
      interval: "weekly"
    open-pull-requests-limit: 10
```

Dependabot creates pull requests for outdated dependencies every week. Review and merge these PRs regularly to stay current.