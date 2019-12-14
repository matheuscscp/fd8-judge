package judge

import (
	"context"
	"fmt"

	"github.com/matheuscscp/fd8-judge/pkg/services"

	"github.com/hashicorp/go-multierror"
)

type (
	// Checker is the program to check the outputs of a programming problem solution and
	// report the results to the given endpoints.
	Checker struct {
		FileService services.FileService

		testCases    []*testCaseFiles
		numTestCases int
	}
)

// Check is the program to check the outputs of a programming problem solution and
// report the results to the given endpoints.
func (e *Checker) Check() error {
	if err := e.prepareBundle(); err != nil {
		return err
	}

	if err := e.preparePrograms(); err != nil {
		return err
	}

	return e.checkTestCases()
}

// prepareBundle downloads and uncompressed the bundle.
func (e *Checker) prepareBundle() error {
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

// preparePrograms compiles the custom checker if present.
func (e *Checker) preparePrograms() error {
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

// checkTestCases loops over the test cases to execute them and upload their outputs.
func (e *Executor) checkTestCases() error {
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
