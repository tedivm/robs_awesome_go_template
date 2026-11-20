package client

// This package provides a type-safe HTTP client generated from the OpenAPI spec.
//
// To generate the client:
//   1. Start the server to generate the OpenAPI spec at /openapi.yaml
//   2. Copy the spec: curl http://localhost:8080/openapi.yaml > ../../docs/openapi.yaml
//   3. Run: go generate ./...
//
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -package client ../../docs/openapi.yaml > client.gen.go
