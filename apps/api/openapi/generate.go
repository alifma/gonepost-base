// Package openapi hosts the go:generate directive for the OpenAPI spec in
// this directory. It has no importable code — it exists only so `go
// generate ./...` (run from apps/api) has a Go file to anchor the command
// to, with this directory as its working directory (so the relative paths
// in codegen.config.yaml resolve correctly).
package openapi

//go:generate go tool oapi-codegen -config codegen.config.yaml openapi.yaml
