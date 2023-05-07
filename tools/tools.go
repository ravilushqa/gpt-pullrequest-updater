//go:build tools

package tools

import (
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
)

//go:generate go build -o ../bin/ github.com/golangci/golangci-lint/cmd/golangci-lint
