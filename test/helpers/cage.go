package helpers

import (
	"os/exec"
	"path/filepath"
)

// ReplaceCagePathAndArgsForTesting receives an encaged command and replaces its path and args
// hijacked by Encage() with values that work for the testing environment.
func ReplaceCagePathAndArgsForTesting(rootRelativePath string, encagedCmd *exec.Cmd) {
	encagedCmd.Path = filepath.Join(rootRelativePath, "bin/fd8-judge")
	encagedCmd.Args = append([]string{encagedCmd.Path}, encagedCmd.Args[1:]...)
}
