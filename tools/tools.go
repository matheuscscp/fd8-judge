// +build tools

package tools

import (
	_ "github.com/golang/mock/mockgen"                      // generates interface mocks
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint" // aggregates linters
	_ "github.com/wadey/gocovmerge"                         // merges coverage files
)
