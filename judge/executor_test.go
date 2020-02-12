// +build integration

package judge_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"testing"

	"github.com/matheuscscp/fd8-judge/judge"
	"github.com/matheuscscp/fd8-judge/pkg/cage"
	"github.com/matheuscscp/fd8-judge/pkg/services/file"
	"github.com/matheuscscp/fd8-judge/pkg/services/program"
	"github.com/matheuscscp/fd8-judge/test/factories"
	"github.com/matheuscscp/fd8-judge/test/fixtures"
	"github.com/stretchr/testify/assert"
)

func TestExecute(t *testing.T) {
	bundleFactory := &factories.ProblemBundleFactory{}
	programFactory := &factories.ProgramFactory{}
	serverFactory := &factories.HTTPServerFactory{}
	serverFiles := &factories.Folder{Name: "serverFiles"}

	listener, server, err := serverFactory.NewFileServer("./serverFiles")
	assert.Equal(t, nil, err)
	port := listener.Addr().(*net.TCPAddr).Port

	var tests = map[string]struct {
		interactor               string
		interactorProgramService string
		solutionProgramService   string
		problemBundle            factories.FileTreeNode
		solutionProgram          string
	}{
		"hello-world-c++11-no-input-file": {
			solutionProgramService: "c++11",
			problemBundle:          fixtures.ProblemBundleHelloWorld(),
			solutionProgram:        fixtures.ProgramCpp11HelloWorld,
		},
		"hello-person-c++11-multiple-input-files": {
			solutionProgramService: "c++11",
			problemBundle:          fixtures.ProblemBundleHelloPerson(),
			solutionProgram:        fixtures.ProgramCpp11HelloPerson,
		},
		"hello-default-interactor-c++11": {
			interactor:             judge.DefaultInteractor,
			solutionProgramService: "c++11",
			problemBundle:          fixtures.ProblemBundleHelloDefaultInteractor(),
			solutionProgram:        fixtures.ProgramCpp11HelloDefaultInteractor,
		},
		"hello-custom-interactor-c++11": {
			interactor:               judge.CustomInteractor,
			interactorProgramService: "c++11",
			solutionProgramService:   "c++11",
			problemBundle:            fixtures.ProblemBundleHelloCustomInteractor(),
			solutionProgram:          fixtures.ProgramCpp11HelloCustomInteractor,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := serverFiles.Write(".")
			assert.Equal(t, nil, err)

			err = bundleFactory.Create(test.problemBundle, "./serverFiles/bundle.tar.gz")
			assert.Equal(t, nil, err)

			err = programFactory.Create(test.solutionProgram, "./serverFiles/solution")
			assert.Equal(t, nil, err)

			var interactorProgramService program.Service
			if test.interactor == judge.CustomInteractor {
				interactorProgramService, err = program.NewService(test.interactorProgramService, nil)
				assert.Equal(t, nil, err)
			}

			solutionProgramService, err := program.NewService(test.solutionProgramService, nil)
			assert.Equal(t, nil, err)

			cage := &factories.Cage{TestCage: cage.New(nil, nil), RootPath: ".."}
			executor := &judge.Executor{
				BundleRequestURL:          fmt.Sprintf("http://localhost:%d/download?path=./serverFiles/bundle.tar.gz", port),
				SolutionRequestURL:        fmt.Sprintf("http://localhost:%d/download?path=./serverFiles/solution", port),
				Interactor:                test.interactor,
				UploadAuthorizedServerURL: fmt.Sprintf("http://localhost:%d/upload", port),
				FileService:               file.NewService(nil),
				InteractorProgramService:  interactorProgramService,
				SolutionProgramService:    solutionProgramService,
				InteractorCage:            cage,
				SolutionCage:              cage,
			}
			err = executor.Execute()
			assert.Equal(t, nil, err)

			problemBundle := test.problemBundle.(*factories.Folder)
			for _, bundleNode := range problemBundle.Children {
				if folder, ok := bundleNode.(*factories.Folder); ok && folder.GetName() == "outputs" {
					for _, outputNode := range folder.Children {
						outputFile := outputNode.(*factories.File)

						outputBytes, err := ioutil.ReadFile("./outputs/" + outputFile.GetName())
						assert.Equal(t, nil, err)
						assert.Equal(t, outputFile.Content, string(outputBytes))

						fileServerOutputBytes, err := ioutil.ReadFile("./serverFiles/" + outputFile.GetName())
						assert.Equal(t, nil, err)
						assert.Equal(t, outputFile.Content, string(fileServerOutputBytes))
					}
					break
				}
			}

			forestToRemove := []factories.FileTreeNode{
				serverFiles,
				&factories.Folder{Name: "bundle"},
				&factories.Folder{Name: "outputs"},
				&factories.File{Name: "solution" + solutionProgramService.GetSourceFileExtension()},
				&factories.File{Name: "solution" + solutionProgramService.GetBinaryFileExtension()},
			}
			if test.interactor == judge.CustomInteractor {
				forestToRemove = append(forestToRemove, &factories.File{
					Name: "interactor" + interactorProgramService.GetSourceFileExtension(),
				}, &factories.File{
					Name: "interactor" + interactorProgramService.GetBinaryFileExtension(),
				})
			}
			for _, tree := range forestToRemove {
				err = tree.Remove(".")
				assert.Equal(t, nil, err)
			}
		})
	}

	err = server.Shutdown(context.Background())
	assert.Equal(t, nil, err)
}
