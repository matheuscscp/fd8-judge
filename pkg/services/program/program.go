package program

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type (
	// Service provides methods to compile and execute programs.
	Service interface {
		// Compile compiles a source code file to a binary file.
		Compile(ctx context.Context, sourceRelativePath, binaryRelativePath string) error

		// GetExecutionCommand returns an *exec.Cmd to execute the given program.
		GetExecutionCommand(ctx context.Context, sourceRelativePath, binaryRelativePath string) *exec.Cmd

		// GetSourceFileExtension returns the extension for source code files names.
		GetSourceFileExtension() string

		// GetBinaryFileExtension returns the extension for binary executable file names.
		GetBinaryFileExtension() string
	}

	serviceRuntime interface {
		Run(cmd *exec.Cmd) error
	}

	defaultServiceRuntime struct {
	}

	// cpp11Service implements compilation and execution for C++ 11.
	cpp11Service struct {
		runtime serviceRuntime
	}
)

// NewService creates a Service according to the given key.
// If nil is passed, the Service will be created with the default serviceRuntime.
func NewService(serviceKey string, runtime serviceRuntime) (Service, error) {
	if runtime == nil {
		runtime = &defaultServiceRuntime{}
	}
	svc, ok := getServices(runtime)[serviceKey]
	if !ok {
		return nil, fmt.Errorf(
			"invalid program service, want one in {%s}, got '%s'",
			strings.Join(GetServices(), ", "),
			serviceKey,
		)
	}
	return svc, nil
}

// GetServices returns a string list of the available program services.
func GetServices() []string {
	programServices := getServices(nil)
	strings := make([]string, 0, len(programServices))
	for key := range programServices {
		strings = append(strings, "'"+key+"'")
	}
	return strings
}

// getServices returns the available program services.
func getServices(runtime serviceRuntime) map[string]Service {
	return map[string]Service{
		"c++11": &cpp11Service{runtime: runtime},
	}
}

func (*defaultServiceRuntime) Run(cmd *exec.Cmd) error {
	return cmd.Run()
}

func (p *cpp11Service) Compile(ctx context.Context, sourceRelativePath, binaryRelativePath string) error {
	cmd := exec.CommandContext(ctx, "g++", "-std=c++11", sourceRelativePath, "-o", binaryRelativePath)
	if err := p.runtime.Run(cmd); err != nil {
		return fmt.Errorf("error compiling for c++11: %w", err)
	}
	return nil
}

func (*cpp11Service) GetExecutionCommand(ctx context.Context, sourceRelativePath, binaryRelativePath string) *exec.Cmd {
	return exec.CommandContext(ctx, binaryRelativePath)
}

func (*cpp11Service) GetSourceFileExtension() string {
	return ".cpp"
}

func (*cpp11Service) GetBinaryFileExtension() string {
	return ""
}
