//go:build tools
// +build tools

package main

import (
	_ "github.com/golang/mock/mockgen"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/goreleaser/goreleaser/v2"
	_ "github.com/mgechev/revive"
	_ "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen"
	_ "honnef.co/go/tools/cmd/staticcheck"
)
