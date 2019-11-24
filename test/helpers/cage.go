package helpers

import (
	"os/exec"
	"path/filepath"
)

// ReplaceCageCommandPathAndArgs receives an encaged command and replaces its path and args
// hijacked by Encage() with values that work for the testing environment.
func ReplaceCageCommandPathAndArgs(rootPath string, encagedCmd *exec.Cmd) {
	encagedCmd.Path = filepath.Join(rootPath, "bin/fd8-judge")
	encagedCmd.Args = append([]string{encagedCmd.Path}, encagedCmd.Args[1:]...)
}
