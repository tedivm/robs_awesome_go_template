#!/usr/bin/env python3
"""Pre-generation hook for Rob's Awesome Go Template.

Validates inputs before project generation.
This script is rendered through Jinja2 by cookiecutter before execution,
so cookiecutter variables are substituted directly.
"""

import sys


def main():
    project_slug = "{{ cookiecutter.project_slug }}"
    import_path = "{{ cookiecutter.import_path }}"
    go_version = "{{ cookiecutter.go_version }}"
    include_web = "{{ cookiecutter.include_web }}"
    include_orm = "{{ cookiecutter.include_orm }}"
    include_cli = "{{ cookiecutter.include_cli }}"
    include_workers = "{{ cookiecutter.include_workers }}"
    project_name = "{{ cookiecutter.project_name }}"

    # Validate project_slug
    if not project_slug:
        print("Error: project_slug cannot be empty.")
        sys.exit(1)

    if not project_slug.replace("_", "").isalnum():
        print(
            "Error: project_slug '"
            + project_slug
            + "' contains invalid characters. Use only lowercase letters, numbers, and underscores."
        )
        sys.exit(1)

    # Validate import_path
    if not import_path:
        print("Error: import_path cannot be empty.")
        sys.exit(1)

    # Validate go_version
    if not go_version:
        print("Error: go_version cannot be empty.")
        sys.exit(1)

    # Check that at least one feature is enabled
    features = [include_web, include_orm, include_cli, include_workers]
    if all(f == "n" for f in features):
        print(
            "Warning: No features selected. The generated project will be essentially empty."
        )

    print("Pre-generation checks passed.")
    print("  Project: " + project_name)
    print("  Slug: " + project_slug)
    print("  Import: " + import_path)


if __name__ == "__main__":
    main()
