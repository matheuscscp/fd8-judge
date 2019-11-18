package services

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type (
	// ProgramService provides methods to compile and execute programs.
	ProgramService interface {
		// Compile compiles a source code file to a binary file.
		Compile(ctx context.Context, sourceRelativePath, binaryRelativePath string) error

		// GetExecutionCommand returns an *exec.Cmd to execute the given program.
		GetExecutionCommand(ctx context.Context, sourceRelativePath, binaryRelativePath string) *exec.Cmd

		// GetSourceFileExtension returns the extension for source code files names.
		GetSourceFileExtension() string

		// GetBinaryFileExtension returns the extension for binary executable file names.
		GetBinaryFileExtension() string
	}

	// ProgramServiceRuntime is the contract to supply for the implementations of ProgramService.
	ProgramServiceRuntime interface {
		// RunCommand runs a command.
		RunCommand(cmd *exec.Cmd) error
	}

	// programServiceDefaultRuntime is the default implementation of ProgramServiceRuntime.
	programServiceDefaultRuntime struct {
	}

	// cpp11ProgramService implements compilation and execution for C++ 11.
	cpp11ProgramService struct {
		runtime ProgramServiceRuntime
	}
)

// NewProgramService creates a ProgramService according to the given key.
// If nil is passed, the ProgramService will be created with the default ProgramServiceRuntime.
func NewProgramService(programServiceKey string, runtime ProgramServiceRuntime) (ProgramService, error) {
	if runtime == nil {
		runtime = &programServiceDefaultRuntime{}
	}
	svc, ok := getProgramServices(runtime)[programServiceKey]
	if !ok {
		return nil, fmt.Errorf(
			"invalid program service, want one in {%s}, got '%s'",
			strings.Join(GetProgramServices(), ", "),
			programServiceKey,
		)
	}
	return svc, nil
}

// GetProgramServices returns a string list of the available program services.
func GetProgramServices() []string {
	programServices := getProgramServices(nil)
	strings := make([]string, 0, len(programServices))
	for key := range programServices {
		strings = append(strings, "'"+key+"'")
	}
	return strings
}

// getProgramServices returns the available program services.
func getProgramServices(runtime ProgramServiceRuntime) map[string]ProgramService {
	return map[string]ProgramService{
		"c++11": &cpp11ProgramService{runtime: runtime},
	}
}

// RunCommand runs a command.
func (*programServiceDefaultRuntime) RunCommand(cmd *exec.Cmd) error {
	return cmd.Run()
}

// Compile compiles a source code file to a binary file.
func (p *cpp11ProgramService) Compile(ctx context.Context, sourceRelativePath, binaryRelativePath string) error {
	cmd := exec.CommandContext(ctx, "g++", "-std=c++11", sourceRelativePath, "-o", binaryRelativePath)
	if err := p.runtime.RunCommand(cmd); err != nil {
		return fmt.Errorf("error compiling for c++11: %w", err)
	}
	return nil
}

// GetExecutionCommand returns an *exec.Cmd to execute the given program.
func (*cpp11ProgramService) GetExecutionCommand(ctx context.Context, sourceRelativePath, binaryRelativePath string) *exec.Cmd {
	return exec.CommandContext(ctx, binaryRelativePath)
}

// GetSourceFileExtension returns the extension for source code files names.
func (*cpp11ProgramService) GetSourceFileExtension() string {
	return ".cpp"
}

// GetBinaryFileExtension returns the extension for binary executable file names.
func (*cpp11ProgramService) GetBinaryFileExtension() string {
	return ""
}
