package cage

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"golang.org/x/sys/unix"
)

type (
	// Cage offers a safe environment to execute a command through a set of restriction options.
	Cage interface {
		// Encage encages the given command, returning a command that will invoke the cage with
		// the arguments necessary to run the given command.
		Encage(monster *exec.Cmd) (*exec.Cmd, error)

		// Execute installs the restrictions in the current process and then executes the command.
		Execute() error
	}

	// DefaultCage is the default implementation for Cage. Uses the golang.org/x/sys/unix package.
	DefaultCage struct {
		// TimeLimit is the maximum time duration for which the process can stay running,
		// before SIGXCPU signal is sent.
		TimeLimit *time.Duration

		// ExecPath is the path to binary/script executable to be executed, and will be passed to
		// unix.Exec() as the (first) argv0 argument.
		ExecPath string

		// ExecArgs are the arguments to be passed to unix.Exec() (through argument argv).
		ExecArgs []string

		runtime defaultCageRuntime
	}

	defaultCageRuntime interface {
		LookPath(file string) (string, error)
		Exec(argv0 string, argv []string, envv []string) error
		Setrlimit(which int, lim *unix.Rlimit) error
	}

	cageDefaultRuntime struct {
	}
)

const (
	// CommandLineCommand is the command-line command used to invoke the cage.
	CommandLineCommand = "cage"

	// CommandLineFlagPrefix is the prefix to be prepended to a flag name to get a valid
	// command-line.
	CommandLineFlagPrefix = "--"

	// TimeLimitFlag is the command line flag for cage to set the TimeLimit option.
	TimeLimitFlag = "time-limit"

	// ExecPathFlag is the command line flag for cage to set the ExecPath property.
	ExecPathFlag = "exec-path"

	// ExecArgsFlag is the command line flag for cage to set the ExecArgs property.
	ExecArgsFlag = "exec-args"
)

// New instantiates a default cage and/or a default runtime and returns them.
func New(cage *DefaultCage, runtime defaultCageRuntime) Cage {
	if cage == nil {
		cage = &DefaultCage{}
	}
	if runtime == nil {
		runtime = &cageDefaultRuntime{}
	}
	cage.runtime = runtime
	return cage
}

// Encage encages the given command, returning a command that will invoke the cage with
// the arguments necessary to run the given command.
func (c *DefaultCage) Encage(monster *exec.Cmd) (*exec.Cmd, error) {
	fd8judge := os.Args[0]
	cagePath := fd8judge
	cageArgs := []string{fd8judge, CommandLineCommand}

	// ensure cagePath is a path with exec.LookPath()
	if filepath.Base(fd8judge) == fd8judge {
		lp, err := c.runtime.LookPath(fd8judge)
		if err != nil {
			return nil, fmt.Errorf("error looking path for fd8-judge: %w", err)
		}
		cagePath = lp
	}

	appendFlag := func(flag, value string) {
		cageArgs = append(cageArgs, CommandLineFlagPrefix+flag, value)
	}

	// append options
	if c.TimeLimit != nil {
		appendFlag(TimeLimitFlag, c.TimeLimit.String())
	}

	// append monster path and args
	appendFlag(ExecPathFlag, monster.Path)
	for _, arg := range monster.Args {
		appendFlag(ExecArgsFlag, arg)
	}

	monster.Path = cagePath
	monster.Args = cageArgs
	return monster, nil
}

// Execute installs the restrictions in the current process and then does the actual unix.Exec().
func (c *DefaultCage) Execute() error {
	if err := c.restrictTimeLimit(); err != nil {
		return err
	}

	if err := c.runtime.Exec(c.ExecPath, c.ExecArgs, nil); err != nil {
		return fmt.Errorf("error in exec syscall: %w", err)
	}
	return nil // never really happens, but go can't guess
}

// restrictTimeLimit enforces the time limit option, setting the maximum amount of time the process
// will be allowed to stay in the CPU.
func (c *DefaultCage) restrictTimeLimit() error {
	if c.TimeLimit != nil {
		timeLimitSeconds := uint64(c.TimeLimit.Seconds())
		if timeLimitSeconds == 0 {
			timeLimitSeconds = 1
		}
		err := c.runtime.Setrlimit(unix.RLIMIT_CPU, &unix.Rlimit{
			Cur: timeLimitSeconds,
			Max: timeLimitSeconds,
		})
		if err != nil {
			return fmt.Errorf("error restricting time limit: %w", err)
		}
	}
	return nil
}

func (*cageDefaultRuntime) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

func (*cageDefaultRuntime) Exec(argv0 string, argv []string, envv []string) error {
	return unix.Exec(argv0, argv, envv)
}

func (*cageDefaultRuntime) Setrlimit(which int, lim *unix.Rlimit) error {
	return unix.Setrlimit(which, lim)
}
