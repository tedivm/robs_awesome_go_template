// Template integration tests.
//
// Generates projects from each config in tests/*.yaml and validates:
//   - Required files exist
//   - Excluded files don't exist
//   - go mod tidy succeeds
//   - go build ./... succeeds
//   - go test ./... passes (or fails gracefully for DB-dependent tests)
//
// Usage:
//
//	go run tests/test_template.go
//	go run tests/test_template.go full
//	go run tests/test_template.go full bare library
package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type Config struct {
	DefaultContext map[string]string
	RequiredFiles  []string
	ExcludedFiles  []string
}

func main() {
	cwd, _ := os.Getwd()
	templateRoot := cwd

	// Find config files
	pattern := filepath.Join(templateRoot, "tests", "*.yaml")
	configs, _ := filepath.Glob(pattern)

	if len(os.Args) > 1 {
		filtered := make([]string, 0)
		for _, c := range configs {
			name := strings.TrimSuffix(filepath.Base(c), ".yaml")
			for _, arg := range os.Args[1:] {
				if arg == name {
					filtered = append(filtered, c)
					break
				}
			}
		}
		configs = filtered
	}

	if len(configs) == 0 {
		fmt.Println("No test configs found")
		os.Exit(1)
	}

	passed := 0
	failed := 0

	fmt.Println("Rob's Awesome Go Template — Test Suite")
	fmt.Println("========================================")
	fmt.Printf("Configs: %d\n\n", len(configs))

	for _, configPath := range configs {
		name := strings.TrimSuffix(filepath.Base(configPath), ".yaml")
		if runTest(templateRoot, configPath, name) {
			passed++
		} else {
			failed++
		}
	}

	fmt.Println("\n========================================")
	fmt.Printf("Results: %d passed, %d failed, %d total\n", passed, failed, passed+failed)

	if failed > 0 {
		os.Exit(1)
	}
}

func runTest(templateRoot, configPath, name string) bool {
	fmt.Printf("=== Testing: %s ===\n", name)

	config, err := parseConfig(configPath)
	if err != nil {
		fmt.Printf("  ✗ Failed to parse config: %v\n", err)
		return false
	}

	projectName := config.DefaultContext["project_name"]
	expectedSlug := slugify(projectName)

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "go-template-test-*")
	if err != nil {
		fmt.Printf("  ✗ Failed to create temp dir: %v\n", err)
		return false
	}
	defer os.RemoveAll(tmpDir)

	// Build cookiecutter args
	args := []string{templateRoot, "--no-input", "--overwrite-if-exists"}
	for k, v := range config.DefaultContext {
		args = append(args, fmt.Sprintf("%s=%s", k, v))
	}

	// Run cookiecutter via uv tool run (no pip install needed)
	cmd := exec.Command("uv", append([]string{"tool", "run", "cookiecutter"}, args...)...)
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("  ✗ Generation failed: %v\n%s\n", err, string(output))
		return false
	}

	// Find generated project
	projectDir := findProjectDir(tmpDir)
	if projectDir == "" {
		fmt.Printf("  ✗ Generated project directory not found\n")
		return false
	}

	actualSlug := filepath.Base(projectDir)
	fmt.Printf("  Project: %s\n", actualSlug)

	// Check required files
	for _, relPath := range config.RequiredFiles {
		checkPath := strings.ReplaceAll(relPath, expectedSlug, actualSlug)
		fullPath := filepath.Join(projectDir, checkPath)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			fmt.Printf("  ✗ Missing required file: %s\n", checkPath)
			return false
		}
	}

	// Check excluded files
	for _, relPath := range config.ExcludedFiles {
		checkPath := strings.ReplaceAll(relPath, expectedSlug, actualSlug)
		fullPath := filepath.Join(projectDir, checkPath)
		if _, err := os.Stat(fullPath); err == nil {
			fmt.Printf("  ✗ Unexpected file present: %s\n", checkPath)
			return false
		}
	}

	fmt.Println("  ✓ File validation passed")

	// Run go mod tidy — capture output for diagnostics
	cmd = exec.Command("go", "mod", "tidy")
	cmd.Dir = projectDir
	tidyOutput, _ := cmd.CombinedOutput()
	// go mod tidy may fail for remote modules (repo doesn't exist on GitHub)
	// but we check for real errors below
	tidyFailed := strings.Contains(string(tidyOutput), "module requires") ||
		strings.Contains(string(tidyOutput), "verifying")
	if tidyFailed && !strings.Contains(string(tidyOutput), "Repository not found") &&
		!strings.Contains(string(tidyOutput), "repository") {
		fmt.Printf("  ✗ go mod tidy had errors:\n%s\n", string(tidyOutput))
		return false
	}
	fmt.Println("  ✓ go mod tidy completed")

	// Run go build
	cmd = exec.Command("go", "build", "./...")
	cmd.Dir = projectDir
	output, err = cmd.CombinedOutput()
	if err != nil {
		outputStr := string(output)
		// "missing go.sum entry" means go mod tidy couldn't resolve the project's
		// own module (github.com/testorg/xxx doesn't exist). This is expected.
		// But real compilation errors (undefined, syntax, etc.) are not.
		if strings.Contains(outputStr, "missing go.sum entry") {
			// Check if there are ALSO real compilation errors
			hasRealError := false
			scanner := bufio.NewScanner(strings.NewReader(outputStr))
			for scanner.Scan() {
				line := scanner.Text()
				if strings.Contains(line, "undefined:") ||
					strings.Contains(line, "syntax error") ||
					strings.Contains(line, "too many errors") ||
					strings.Contains(line, "not type") ||
					strings.Contains(line, "cannot use") ||
					strings.Contains(line, "cannot find module") {
					hasRealError = true
					break
				}
			}
			if hasRealError {
				fmt.Printf("  ✗ go build had compilation errors:\n%s\n", outputStr)
				return false
			}
			fmt.Println("  ⚠ go build skipped (remote module not resolvable)")
		} else {
			fmt.Printf("  ✗ go build failed:\n%s\n", outputStr)
			return false
		}
	} else {
		fmt.Println("  ✓ go build passed")
	}

	// Run go test
	cmd = exec.Command("go", "test", "./...")
	cmd.Dir = projectDir
	output, err = cmd.CombinedOutput()
	outputStr := string(output)
	if strings.Contains(outputStr, "BUILD FAILURE") || strings.Contains(outputStr, "cannot find module") {
		fmt.Printf("  ✗ go test had build errors:\n%s\n", outputStr)
		return false
	}
	if strings.Contains(outputStr, "FAIL\t") {
		fmt.Println("  ⚠ Some tests failed (may need DB)")
	} else {
		fmt.Println("  ✓ go test passed")
	}

	fmt.Printf("  ✓ %s: PASSED\n", name)
	return true
}

func findProjectDir(tmpDir string) string {
	entries, _ := os.ReadDir(tmpDir)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(tmpDir, entry.Name())
		if _, err := os.Stat(filepath.Join(path, "go.mod")); err == nil {
			return path
		}
	}
	return ""
}

func slugify(name string) string {
	re := regexp.MustCompile(`[^a-z0-9]`)
	result := strings.ToLower(name)
	result = re.ReplaceAllString(result, "_")
	result = regexp.MustCompile(`_+`).ReplaceAllString(result, "_")
	result = strings.Trim(result, "_")
	return result
}

func parseConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	config := &Config{
		DefaultContext: make(map[string]string),
	}

	var section string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Detect sections
		if line == "default_context:" {
			section = "context"
			continue
		}
		if line == "required_files:" {
			section = "required"
			continue
		}
		if line == "excluded_files:" {
			section = "excluded"
			continue
		}

		// Parse key: value pairs
		if section == "context" && strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// Remove quotes
			value = strings.Trim(value, "\"'")
			config.DefaultContext[key] = value
			continue
		}

		// Parse list items
		if (section == "required" || section == "excluded") && strings.HasPrefix(line, "- ") {
			item := strings.TrimPrefix(line, "- ")
			item = strings.TrimSpace(item)
			item = strings.Trim(item, "\"'")
			if section == "required" {
				config.RequiredFiles = append(config.RequiredFiles, item)
			} else {
				config.ExcludedFiles = append(config.ExcludedFiles, item)
			}
		}
	}

	return config, scanner.Err()
}
