# Rob's Awesome Go Template — Development tasks
# Run 'just' to see available tasks

test:
	@echo "Running template tests..."
	go run tests/test_template.go

test-verbose:
	@echo "Running template tests (verbose)..."
	go run tests/test_template.go $(CONFIG)

test-%:
	@echo "Running template test: $*"
	go run tests/test_template.go $*

lint:
	@echo "Linting template files..."
	ruff check hooks/ tests/

fmt:
	@echo "Formatting template files..."
	gofmt -w -s tests/
	ruff format hooks/ tests/

validate:
	@echo "Validating cookiecutter.json..."
	python3 -c "import json; json.load(open('cookiecutter.json'))" && echo "✓ cookiecutter.json is valid JSON"

all: validate lint test
