package judge

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/matheuscscp/fd8-judge/pkg/services/file"
)

type testCaseFiles struct {
	input, output string
}

func listTestCases(fileSvc file.Service) ([]*testCaseFiles, error) {
	inputFiles, err := fileSvc.ListFiles(folderPathBundleInputs)
	if err != nil && !errors.Is(err, &file.NoSuchFolderError{}) {
		return nil, fmt.Errorf("error listing input files: %w", err)
	}
	testCases := make([]*testCaseFiles, 1)
	if len(inputFiles) == 0 {
		testCases[0] = &testCaseFiles{
			input:  devNull,
			output: filePathSolutionSingleOutput,
		}
	} else {
		testCases = make([]*testCaseFiles, len(inputFiles))
		for i := range inputFiles {
			testCases[i] = &testCaseFiles{
				input:  filepath.Join(folderPathBundleInputs, inputFiles[i]),
				output: filepath.Join(folderPathSolutionOutputs, inputFiles[i]),
			}
		}
	}
	return testCases, nil
}
