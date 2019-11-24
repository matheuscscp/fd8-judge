package factories

import (
	"os/exec"

	"github.com/matheuscscp/fd8-judge/pkg/cage"
	"github.com/matheuscscp/fd8-judge/test/helpers"
)

// Cage wraps a test cage.Cage and replaces the command's path and args hijacked by Encage() with
// values appropriate for tests.
type Cage struct {
	// TestCage is the wrapped cage.
	TestCage cage.Cage

	// RootPath should be a path to the root folder of the project where the helper function used
	// to replace the hijacked values can properly accomplish its work.
	RootPath string
}

// Encage fixes command's path and args hijacked by TestCage.Encage().
func (c *Cage) Encage(monster *exec.Cmd) *exec.Cmd {
	cmd := c.TestCage.Encage(monster)
	helpers.ReplaceCageCommandPathAndArgs(c.RootPath, cmd)
	return cmd
}

// Execute delegates to testCage.Encage().
func (c *Cage) Execute() error {
	return c.TestCage.Execute()
}
