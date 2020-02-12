package judge

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/matheuscscp/fd8-judge/pkg/cage"
	"github.com/matheuscscp/fd8-judge/pkg/services/file"
	"github.com/matheuscscp/fd8-judge/pkg/services/program"
)

type (
	// Executor is the program to execute a programming problem solution feeding it with a set of
	// tests and to upload the results to the given endpoints.
	Executor struct {
		// BundleRequestURL is the URL endpoint to download the problem bundle.
		BundleRequestURL string

		// BundleRequestHeaders are the HTTP headers to be sent when downloading the problem bundle.
		BundleRequestHeaders http.Header

		// SolutionRequestURL is the URL endpoint to download the problem solution.
		SolutionRequestURL string

		// SolutionRequestHeaders are the HTTP headers to be sent when downloading the problem solution.
		SolutionRequestHeaders http.Header

		// Interactor identifies the execution style of the problem solution, defining the interactor
		// program that should feed input and collect output to/from the solution process, where the
		// values allowed are NoInteractor (just feed the input file and collect the output),
		// DefaultInteractor (read a line from the input file with the number of following lines to read
		// and feed to the solution process, then read a line from the solution process output with the
		// number of following lines to read and store, then start again), or a path to an interactor
		// source code within the uncompressed bundle.
		Interactor string

		// UploadAuthorizedServerURL is the URL endpoint to get an authorized upload request for the
		// compressed outputs of the problem solution.
		UploadAuthorizedServerURL string

		// FileService offers the necessary file operations for Executor.
		FileService file.Service

		// InteractorProgramService offers methods to compile and execute the interactor.
		InteractorProgramService program.Service

		// SolutionProgramService offers methods to compile and execute the solution.
		SolutionProgramService program.Service

		// InteractorCage restricts the interactor process.
		InteractorCage cage.Cage

		// SolutionCage restricts the solution process.
		SolutionCage cage.Cage

		testCases                []*testCaseFiles
		numTestCases             int
		filePathInteractorSource string
		filePathInteractorBinary string
		filePathSolutionSource   string
		filePathSolutionBinary   string
	}
)

// Execute is the program to execute a programming problem solution feeding it with a set of
// tests and to upload the results to the given endpoints.
func (e *Executor) Execute() error {
	if err := e.prepareBundle(); err != nil {
		return err
	}

	if err := e.preparePrograms(); err != nil {
		return err
	}

	return e.executeTestCases()
}

// prepareBundle downloads the bundle, uncompresses it and removes undesired files.
func (e *Executor) prepareBundle() error {
	if _, err := e.FileService.DownloadFile(filePathCompressedBundle, e.BundleRequestURL, e.BundleRequestHeaders); err != nil {
		return fmt.Errorf("error downloading compressed problem bundle for execution: %w", err)
	}

	if err := e.FileService.Uncompress(filePathCompressedBundle, "."); err != nil {
		return fmt.Errorf("error uncompressing problem bundle for execution: %w", err)
	}

	if err := e.FileService.RemoveFileTree(filePathCompressedBundle); err != nil {
		return fmt.Errorf("error removing compressed problem bundle for execution: %w", err)
	}

	if err := e.FileService.RemoveFileTree(folderPathBundleOutputs); err != nil {
		return fmt.Errorf("error removing correct outputs for execution: %w", err)
	}

	var err error
	e.testCases, err = listTestCases(e.FileService)
	if err != nil {
		return fmt.Errorf("error listing test cases: %w", err)
	}
	e.numTestCases = len(e.testCases)

	return nil
}

// preparePrograms downloads the solution and compiles it together with the custom interactor if
// present.
func (e *Executor) preparePrograms() error {
	e.filePathSolutionSource = filesPathPrefixSolution + e.SolutionProgramService.GetSourceFileExtension()
	if _, err := e.FileService.DownloadFile(e.filePathSolutionSource, e.SolutionRequestURL, e.SolutionRequestHeaders); err != nil {
		return fmt.Errorf("error downloading problem solution: %w", err)
	}

	e.filePathSolutionBinary = filesPathPrefixSolution + e.SolutionProgramService.GetBinaryFileExtension()
	if err := e.SolutionProgramService.Compile(context.TODO(), e.filePathSolutionSource, e.filePathSolutionBinary); err != nil {
		return fmt.Errorf("error compiling problem solution: %w", err)
	}

	if e.InteractorProgramService != nil {
		e.filePathInteractorSource = filesPathPrefixInteractor + e.InteractorProgramService.GetSourceFileExtension()
		if err := e.FileService.MoveFileTree(filePathBundleInteractor, e.filePathInteractorSource); err != nil {
			return fmt.Errorf("error moving interactor source code: %w", err)
		}

		e.filePathInteractorBinary = filesPathPrefixInteractor + e.InteractorProgramService.GetBinaryFileExtension()
		if err := e.InteractorProgramService.Compile(context.TODO(), e.filePathInteractorSource, e.filePathInteractorBinary); err != nil {
			return fmt.Errorf("error compiling custom interactor: %w", err)
		}
	}

	return nil
}

// executeTestCases loops over the test cases to execute them and upload their outputs.
func (e *Executor) executeTestCases() error {
	runFunc, err := e.getRunFunction()
	if err != nil {
		return err
	}

	// upload outputs in a worker
	jobs := make(chan string, e.numTestCases)
	done := make(chan error, 1)
	go func() {
		for i := 0; i < e.numTestCases; i++ {
			testCaseOutput, more := <-jobs
			if !more {
				done <- nil
				return
			}
			if err := e.FileService.UploadFile(testCaseOutput, e.UploadAuthorizedServerURL); err != nil {
				done <- fmt.Errorf("error uploading test case output: %w", err)
				return
			}
		}

		done <- nil
	}()

	for _, testCase := range e.testCases {
		if err := runFunc(testCase); err != nil {
			close(jobs)
			if jobErr := <-done; err != nil {
				return multierror.Append(err, jobErr)
			}
			return err
		}

		jobs <- testCase.output
	}

	return <-done
}

// getRunFunction reads e.Interactor to return the appropriate run function.
func (e *Executor) getRunFunction() (func(testCase *testCaseFiles) error, error) {
	switch e.Interactor {
	case NoInteractor:
		return e.runWithoutInteractor, nil
	case DefaultInteractor:
		return e.runWithDefaultInteractor, nil
	case CustomInteractor:
		return e.runWithCustomInteractor, nil
	default:
		return nil, fmt.Errorf(
			"invalid interactor, got '%s', want one in {%s}", e.Interactor,
			strings.Join([]string{NoInteractor, DefaultInteractor, CustomInteractor}, ", "),
		)
	}
}

// runWithoutInteractor executes the problem solution without interactor (only feeds input and
// stores output).
func (e *Executor) runWithoutInteractor(testCase *testCaseFiles) error {
	var err error

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	solution := e.SolutionProgramService.GetExecutionCommand(ctx, e.filePathSolutionSource, e.filePathSolutionBinary)
	solution, err = e.SolutionCage.Encage(solution)
	if err != nil {
		return fmt.Errorf("error encaging solution command: %w", err)
	}

	input, err := e.FileService.OpenFile(testCase.input)
	if err != nil {
		return fmt.Errorf("error opening input file for test case execution: %w", err)
	}
	defer input.Close()

	output, err := e.FileService.CreateFile(testCase.output)
	if err != nil {
		return fmt.Errorf("error creating output file for test case execution: %w", err)
	}
	defer output.Close()

	solutionInput, err := solution.StdinPipe()
	if err != nil {
		return fmt.Errorf("error creating pipe for solution input: %w", err)
	}
	defer solutionInput.Close()

	solutionOutput, err := solution.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error creating pipe for solution output: %w", err)
	}
	defer solutionOutput.Close()

	if err := solution.Start(); err != nil {
		return fmt.Errorf("error starting problem solution process: %w", err)
	}
	defer func() { _ = solution.Wait() }()

	pipeErrors := make(chan error, 2)
	go func() {
		if _, err := io.Copy(solutionInput, input); err != nil {
			pipeErrors <- fmt.Errorf("error copying input file to solution input pipe: %w", err)
		} else {
			pipeErrors <- solutionInput.Close() // tells solution process to exit
		}
	}()
	go func() {
		if _, err := io.Copy(output, solutionOutput); err != nil {
			pipeErrors <- fmt.Errorf("error copying solution output pipe to output file: %w", err)
		} else {
			pipeErrors <- nil
		}
	}()
	for i := 0; i < 2; i++ {
		pipeError := <-pipeErrors
		if pipeError != nil {
			return pipeError
		}
	}

	return nil
}

// runWithDefaultInteractor executes the problem solution with the default interactor.
func (e *Executor) runWithDefaultInteractor(testCase *testCaseFiles) error {
	var err error

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	solution := e.SolutionProgramService.GetExecutionCommand(ctx, e.filePathSolutionSource, e.filePathSolutionBinary)
	solution, err = e.SolutionCage.Encage(solution)
	if err != nil {
		return fmt.Errorf("error encaging solution command: %w", err)
	}

	input, err := e.FileService.OpenFile(testCase.input)
	if err != nil {
		return fmt.Errorf("error opening input file for test case execution: %w", err)
	}
	defer input.Close()

	output, err := e.FileService.CreateFile(testCase.output)
	if err != nil {
		return fmt.Errorf("error creating output file for test case execution: %w", err)
	}
	defer output.Close()

	solutionInput, err := solution.StdinPipe()
	if err != nil {
		return fmt.Errorf("error creating pipe for solution input: %w", err)
	}
	defer solutionInput.Close()

	solutionOutput, err := solution.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error creating pipe for solution output: %w", err)
	}
	defer solutionOutput.Close()

	if err := solution.Start(); err != nil {
		return fmt.Errorf("error starting problem solution process: %w", err)
	}
	defer func() { _ = solution.Wait() }()

	inputScanner := bufio.NewScanner(input)
	outputScanner := bufio.NewScanner(solutionOutput)
	for inputScanner.Scan() {
		var linesToFeed uint
		if _, err := fmt.Sscanf(inputScanner.Text(), "%d", &linesToFeed); err != nil {
			return fmt.Errorf("error scanning lines to feed for interactive input: %w", err)
		}

		for i := uint(0); i < linesToFeed; i++ {
			if !inputScanner.Scan() {
				return fmt.Errorf("error reading line from test case input: %w", err)
			}

			line := inputScanner.Text() + "\n"
			if _, err := solutionInput.Write([]byte(line)); err != nil {
				return fmt.Errorf("error writing line to solution process stdin: %w", err)
			}
		}

		var linesToCollect uint
		if !outputScanner.Scan() {
			return fmt.Errorf("problem solution stopped responding interactor: %w", err)
		}
		if _, err := fmt.Sscanf(outputScanner.Text(), "%d", &linesToCollect); err != nil {
			return fmt.Errorf("error scanning lines to collect for interactive output: %w", err)
		}

		for i := uint(0); i < linesToCollect; i++ {
			if !outputScanner.Scan() {
				return fmt.Errorf("problem solution stopped responding interactor: %w", err)
			}

			line := outputScanner.Text() + "\n"
			if _, err := output.Write([]byte(line)); err != nil {
				return fmt.Errorf("error writing problem solution output: %w", err)
			}
		}
	}
	solutionInput.Close() // tells solution process to exit

	return nil
}

// runWithCustomInteractor executes the problem solution with the given custom interactor.
func (e *Executor) runWithCustomInteractor(testCase *testCaseFiles) error {
	var err error

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	interactor := e.InteractorProgramService.GetExecutionCommand(ctx, e.filePathInteractorSource, e.filePathInteractorBinary)
	interactor, err = e.InteractorCage.Encage(interactor)
	if err != nil {
		return fmt.Errorf("error encaging interactor command: %w", err)
	}

	solution := e.SolutionProgramService.GetExecutionCommand(ctx, e.filePathSolutionSource, e.filePathSolutionBinary)
	solution, err = e.SolutionCage.Encage(solution)
	if err != nil {
		return fmt.Errorf("error encaging solution command: %w", err)
	}

	output, err := e.FileService.CreateFile(testCase.output)
	if err != nil {
		return fmt.Errorf("error creating output file for test case execution: %w", err)
	}
	output.Close()

	interactorInput, err := interactor.StdinPipe()
	if err != nil {
		return fmt.Errorf("error creating pipe for interactor input: %w", err)
	}
	defer interactorInput.Close()

	interactorOutput, err := interactor.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error creating pipe for interactor output: %w", err)
	}
	defer interactorOutput.Close()

	solutionInput, err := solution.StdinPipe()
	if err != nil {
		return fmt.Errorf("error creating pipe for solution input: %w", err)
	}
	defer solutionInput.Close()

	solutionOutput, err := solution.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error creating pipe for solution output: %w", err)
	}
	defer solutionOutput.Close()

	if err := interactor.Start(); err != nil {
		return fmt.Errorf("error starting interactor process: %w", err)
	}
	defer func() { _ = interactor.Wait() }()

	if err := solution.Start(); err != nil {
		return fmt.Errorf("error starting solution process: %w", err)
	}
	defer func() { _ = solution.Wait() }()

	_, err = interactorInput.Write([]byte(testCase.input + "\n" + testCase.output + "\n"))
	if err != nil {
		return fmt.Errorf("error writing input and output file paths to interactor input: %w", err)
	}

	pipeErrors := make(chan error, 2)
	go func() {
		if _, err := io.Copy(solutionInput, interactorOutput); err != nil {
			pipeErrors <- fmt.Errorf("error copying interactor output pipe to solution input pipe: %w", err)
		} else {
			pipeErrors <- solutionInput.Close() // tells solution process to exit
		}
	}()
	go func() {
		if _, err := io.Copy(interactorInput, solutionOutput); err != nil {
			pipeErrors <- fmt.Errorf("error copying solution output pipe to interactor input pipe: %w", err)
		} else {
			pipeErrors <- nil
		}
	}()
	for i := 0; i < 2; i++ {
		pipeError := <-pipeErrors
		if pipeError != nil {
			return pipeError
		}
	}

	return nil
}
