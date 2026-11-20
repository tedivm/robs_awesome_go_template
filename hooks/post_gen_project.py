#!/usr/bin/env python3
"""Post-generation hook for Rob's Awesome Go Template."""

import os
import sys
import shutil
import subprocess


def remove_path(path):
    if os.path.exists(path):
        if os.path.isdir(path):
            shutil.rmtree(path)
            print("  Removed directory: " + path)
        else:
            os.remove(path)
            print("  Removed file: " + path)


def get_bool(value):
    return str(value).lower() == "y"


def main():
    include_web = get_bool("{{ cookiecutter.include_web }}")
    include_orm = get_bool("{{ cookiecutter.include_orm }}")
    include_cli = get_bool("{{ cookiecutter.include_cli }}")
    include_workers = get_bool("{{ cookiecutter.include_workers }}")
    include_docker = get_bool("{{ cookiecutter.include_docker }}")
    include_ai = get_bool("{{ cookiecutter.include_ai }}")
    license_type = "{{ cookiecutter.license }}"
    project_name = "{{ cookiecutter.project_name }}"
    project_slug = "{{ cookiecutter.project_slug }}"

    print("\n=== Post-generation cleanup ===\n")

    if not include_web:
        print("[web disabled] Removing web-related files...")
        remove_path("cmd/" + project_slug)
        remove_path("internal/handler")
        remove_path("internal/renderer")
        remove_path("internal/client")
        remove_path("internal/middleware")
        remove_path("static")
        remove_path("air.toml")
        remove_path("docs/dev/api.md")
        remove_path("test/handler")
        remove_path("test/middleware")
        remove_path("test/template")
        if not include_cli:
            remove_path(".github/workflows/release.yaml")
            remove_path(".goreleaser.yaml")
        if include_ai:
            remove_path(".agents/skills/adding-api-endpoints")
            remove_path(".agents/skills/adding-new-entities")

    if not include_orm:
        print("[orm disabled] Removing ORM-related files...")
        remove_path("internal/model")
        remove_path("internal/service")
        remove_path("migrations")
        remove_path("docs/dev/database.md")
        remove_path("docs/dev/migrations.md")
        remove_path("test/util/testdb.go")
        remove_path("test/service")
        remove_path("test/util")
        remove_path(".github/workflows/integration.yaml")
        remove_path("internal/handler/items.go")
        remove_path("internal/handler/auth.go")
        remove_path("test/handler/health_test.go")
        remove_path("test/handler/items_test.go")
        if include_ai:
            remove_path(".agents/skills/creating-bun-models")
            remove_path(".agents/skills/managing-migrations")

    if not include_cli or not include_orm:
        if not include_cli:
            print("[cli disabled] Removing CLI-related files...")
        elif not include_orm:
            print("[cli requires orm] Removing CLI-related files (ORM is disabled)...")
        remove_path("cmd/cli")
        remove_path("internal/cli")
        remove_path("docs/dev/cli.md")
        if not include_web:
            remove_path(".github/workflows/release.yaml")
            remove_path(".goreleaser.yaml")
        if include_ai:
            remove_path(".agents/skills/adding-cli-commands")

    if not include_workers:
        print("[workers disabled] Removing worker-related files...")
        remove_path("internal/worker")
        remove_path("cmd/worker")
        remove_path("docs/dev/workers.md")
        remove_path("test/worker")
        if include_ai:
            remove_path(".agents/skills/adding-background-tasks")

    if not include_docker or (
        not include_web and not include_cli and not include_workers
    ):
        if not include_docker:
            print("[docker disabled] Removing Docker-related files...")
        else:
            print(
                "[docker has no target] Removing Docker-related files (no web, cli, or workers)..."
            )
        remove_path("Dockerfile")
        remove_path("docker-compose.yaml")
        remove_path("docker-compose.prod.yaml")
        remove_path(".dockerignore")
        remove_path("scripts/dev.sh")
        remove_path(".github/workflows/docker.yaml")
        remove_path("docs/dev/docker.md")
        if include_ai:
            remove_path(".agents/skills/running-docker-services")

    # Remove prod compose if docker is enabled but there's no web or worker service
    if include_docker and not include_web and not include_workers:
        print(
            "[docker has no prod services] Removing production compose file (no web or workers)..."
        )
        remove_path("docker-compose.prod.yaml")

    if not include_ai:
        print("[ai disabled] Removing AI-related files...")
        remove_path("AGENTS.md")
        remove_path(".agents")

    if license_type == "None":
        print("[no license] Removing LICENSE...")
        remove_path("LICENSE")

    # Remove empty parent directories
    dirs_to_check = [
        "cmd",
        "internal",
        "docs/dev",
        "test",
        "scripts",
        ".agents/skills",
        ".github/workflows",
    ]
    for directory in dirs_to_check:
        for root, dirs, files in os.walk(directory, topdown=False):
            if not os.path.exists(root):
                continue
            if not os.listdir(root):
                try:
                    os.rmdir(root)
                    print("  Removed empty directory: " + root)
                except OSError:
                    pass

    # Run go mod tidy
    print("\n=== Running go mod tidy ===")
    try:
        result = subprocess.run(
            ["go", "mod", "tidy"], capture_output=True, text=True, timeout=120
        )
        if result.returncode != 0:
            print("Warning: go mod tidy failed:")
            print(result.stderr)
        else:
            print("go mod tidy completed successfully.")
    except FileNotFoundError:
        print("Warning: 'go' not found. Skipping go mod tidy.")
    except subprocess.TimeoutExpired:
        print("Warning: go mod tidy timed out.")

    # Run gofmt
    print("\n=== Formatting with gofmt ===")
    try:
        result = subprocess.run(
            ["gofmt", "-w", "-s", "."], capture_output=True, text=True, timeout=60
        )
        if result.returncode != 0:
            print("Warning: gofmt failed:")
            print(result.stderr)
        else:
            print("gofmt completed successfully.")
    except FileNotFoundError:
        print("Warning: 'gofmt' not found. Skipping formatting.")

    yn = lambda v: "Yes" if v else "No"
    print("\n=== Generation complete: " + project_name + " ===")
    print("Slug: " + project_slug)
    print(
        "Web: "
        + yn(include_web)
        + " | ORM: "
        + yn(include_orm)
        + " | CLI: "
        + yn(include_cli)
        + " | Workers: "
        + yn(include_workers)
    )
    print("Docker: " + yn(include_docker) + " | AI: " + yn(include_ai) + "\n")


if __name__ == "__main__":
    main()
