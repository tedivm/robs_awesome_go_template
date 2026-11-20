//go:build tools

package tools

import (
{% if cookiecutter.include_web == "y" %}
	_ "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen"
{% endif %}
)
