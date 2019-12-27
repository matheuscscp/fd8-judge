package main

import (
	"github.com/matheuscscp/fd8-judge/cmd"
	_ "github.com/matheuscscp/fd8-judge/cmd/api" // force package initialization
)

func main() {
	cmd.Execute()
}
