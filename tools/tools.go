// +build tools

package tools

import (
	_ "github.com/golang/mock/mockgen"                                 // generates interface mocks
	_ "github.com/golang/protobuf/protoc-gen-go"                       // protoc-gen-go
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"            // aggregates linters
	_ "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway" // protoc-gen-grpc-gateway
	_ "github.com/uber/prototool/cmd/prototool"                        // swiss army knife for protocol buffers
	_ "github.com/wadey/gocovmerge"                                    // merges coverage files
)
