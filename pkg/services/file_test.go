// +build integration

package services_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/matheuscscp/fd8-judge/pkg/services"
	"github.com/matheuscscp/fd8-judge/test/factories"
	"github.com/matheuscscp/fd8-judge/test/fixtures"
	"github.com/stretchr/testify/assert"
)

func TestDownloadFile(t *testing.T) {
	// create server
	serverFactory := &factories.HTTPServerFactory{}
	listener, server, err := serverFactory.NewDummy()
	assert.Equal(t, nil, err)
	port := listener.Addr().(*net.TCPAddr).Port

	// download bytes
	const (
		relativePath        = "./TestDownloadFile.tmp"
		payload             = "PAYLOAD"
		bytesToBeDownloaded = int64(len(payload))
	)
	fileSvc := services.NewFileService(nil)
	bytesDownloaded, err := fileSvc.DownloadFile(
		relativePath,
		fmt.Sprintf("http://localhost:%d/dummy", port),
		nil,
	)
	assert.Equal(t, nil, err)
	assert.Equal(t, bytesToBeDownloaded, bytesDownloaded)

	// check downloaded content
	downloadedContentBytes, err := ioutil.ReadFile(relativePath)
	assert.Equal(t, nil, err)
	assert.Equal(t, []byte(payload), downloadedContentBytes)

	// erase file
	err = os.Remove(relativePath)
	assert.Equal(t, nil, err)

	// shutdown test server
	err = server.Shutdown(context.Background())
	assert.Equal(t, nil, err)
}

func TestUploadFile(t *testing.T) {
	// create server
	serverFactory := &factories.HTTPServerFactory{}
	listener, server, err := serverFactory.NewDummyUploader()
	assert.Equal(t, nil, err)
	port := listener.Addr().(*net.TCPAddr).Port
	authorizedServerURL := fmt.Sprintf("http://localhost:%d/upload-info", port)
	objectURL := fmt.Sprintf("http://localhost:%d/upload", port)

	// create file
	const (
		relativePath      = "./TestUploadFile.tmp"
		payload           = "PAYLOAD"
		bytesToBeUploaded = len(payload)
	)
	err = ioutil.WriteFile(relativePath, []byte(payload), os.ModePerm)
	assert.Equal(t, nil, err)

	// upload
	fileSvc := services.NewFileService(nil)
	err = fileSvc.UploadFile(relativePath, authorizedServerURL)
	assert.Equal(t, nil, err)

	// check uploaded content
	resp, err := http.Get(objectURL)
	assert.Equal(t, nil, err)
	defer resp.Body.Close()
	uploadedContentBytes, err := ioutil.ReadAll(resp.Body)
	assert.Equal(t, nil, err)
	assert.Equal(t, []byte(payload), uploadedContentBytes)

	// erase file
	err = os.Remove(relativePath)
	assert.Equal(t, nil, err)

	// shutdown test server
	err = server.Shutdown(context.Background())
	assert.Equal(t, nil, err)
}

func TestCompressAndUncompress(t *testing.T) {
	// cannot run these tests in parallel because they mess with the file system

	fileSvc := services.NewFileService(nil)

	var tests = map[string]struct {
		fixture                  factories.FileTreeNode
		expectedFileTree         factories.FileTreeNode
		inputRelativePath        string
		uncompressionRootPath    string
		uncompressedRelativePath string
	}{
		"single-file": {
			fixture:                  fixtures.SingleFile(),
			expectedFileTree:         fixtures.SingleFile(),
			inputRelativePath:        "./SingleFile.txt",
			uncompressionRootPath:    ".",
			uncompressedRelativePath: "./SingleFile.txt",
		},
		"empty-folder": {
			fixture:                  fixtures.EmptyFolder(),
			expectedFileTree:         fixtures.EmptyFolder(),
			inputRelativePath:        "./EmptyFolder",
			uncompressionRootPath:    ".",
			uncompressedRelativePath: "./EmptyFolder",
		},
		"big-folder": {
			fixture:                  fixtures.TestFolder(),
			expectedFileTree:         fixtures.TestFolder(),
			inputRelativePath:        "./TestFolder",
			uncompressionRootPath:    ".",
			uncompressedRelativePath: "./TestFolder",
		},
		"cut-path-before-single-file": {
			fixture:                  fixtures.TestFolderOneFile(),
			expectedFileTree:         fixtures.SingleFile(),
			inputRelativePath:        "./TestFolderOneFile/SingleFile.txt",
			uncompressionRootPath:    ".",
			uncompressedRelativePath: "./SingleFile.txt",
		},
		"cut-path-before-empty-folder": {
			fixture:                  fixtures.TestFolderOneFolder(),
			expectedFileTree:         fixtures.EmptyFolder(),
			inputRelativePath:        "./TestFolderOneFolder/EmptyFolder",
			uncompressionRootPath:    ".",
			uncompressedRelativePath: "./EmptyFolder",
		},
		"cut-long-path-before-big-folder": {
			fixture:                  fixtures.TestDummyRootFolder(),
			expectedFileTree:         fixtures.TestFolder(),
			inputRelativePath:        "./TestDummyRootFolder//MiddleFolder/TestFolder",
			uncompressionRootPath:    ".",
			uncompressedRelativePath: "./TestFolder",
		},
		"non-dot-uncompression-root-path": {
			fixture:                  fixtures.TestDummyRootFolder(),
			expectedFileTree:         fixtures.TestFolder(),
			inputRelativePath:        "./TestDummyRootFolder//MiddleFolder/TestFolder",
			uncompressionRootPath:    "./rootFolder",
			uncompressedRelativePath: "./rootFolder/TestFolder",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := test.fixture.Write(".")
			assert.Equal(t, nil, err)

			err = fileSvc.Compress(test.inputRelativePath, "./TestCompressedFile.tar.gz")
			assert.Equal(t, nil, err)

			err = test.fixture.Remove(".")
			assert.Equal(t, nil, err)

			err = fileSvc.Uncompress("./TestCompressedFile.tar.gz", test.uncompressionRootPath)
			assert.Equal(t, nil, err)

			err = os.Remove("./TestCompressedFile.tar.gz")
			assert.Equal(t, nil, err)

			fileTree, err := factories.ReadFileTree(test.uncompressedRelativePath, true)
			assert.Equal(t, nil, err)
			assert.Equal(t, test.expectedFileTree, fileTree)

			if filepath.Clean(test.uncompressionRootPath) == "." {
				err = fileTree.Remove(test.uncompressionRootPath)
			} else {
				err = os.RemoveAll(test.uncompressionRootPath)
			}
			assert.Equal(t, nil, err)
		})
	}
}

func TestRemoveFileTree(t *testing.T) {
	fileSvc := services.NewFileService(nil)
	fixture := fixtures.TestFolder()

	err := fixture.Write(".")
	assert.Equal(t, nil, err)

	fileTree, err := factories.ReadFileTree(fixture.GetName(), true)
	assert.Equal(t, nil, err)
	assert.Equal(t, fixture, fileTree)

	err = fileSvc.RemoveFileTree("./" + fixture.GetName())
	assert.Equal(t, nil, err)

	_, err = factories.ReadFileTree(fixture.GetName(), true)
	assert.Equal(t, true, strings.Contains(err.Error(), "error Stat()ing relative path to read file tree"))
}

func TestOpenFile(t *testing.T) {
	fileSvc := services.NewFileService(nil)
	fixture, ok := fixtures.SingleFile().(*factories.File)
	assert.Equal(t, true, ok)

	err := fixture.Write(".")
	assert.Equal(t, nil, err)

	file, err := fileSvc.OpenFile("./" + fixture.GetName())
	assert.Equal(t, nil, err)

	bytes, err := ioutil.ReadAll(file)
	assert.Equal(t, nil, err)
	assert.Equal(t, []byte(fixture.Content), bytes)

	err = file.Close()
	assert.Equal(t, nil, err)

	err = fixture.Remove(".")
	assert.Equal(t, nil, err)
}

func TestCreateFile(t *testing.T) {
	fileSvc := services.NewFileService(nil)
	fixture, ok := fixtures.SingleFile().(*factories.File)
	assert.Equal(t, true, ok)

	file, err := fileSvc.CreateFile("./" + fixture.GetName())
	assert.Equal(t, nil, err)

	numBytes, err := file.Write([]byte(fixture.Content))
	assert.Equal(t, nil, err)
	assert.Equal(t, len(fixture.Content), numBytes)

	err = file.Close()
	assert.Equal(t, nil, err)

	bytes, err := ioutil.ReadFile("./" + fixture.GetName())
	assert.Equal(t, nil, err)
	assert.Equal(t, []byte(fixture.Content), bytes)

	err = fixture.Remove(".")
	assert.Equal(t, nil, err)
}

func TestListFiles(t *testing.T) {
	fileSvc := services.NewFileService(nil)
	fixture, ok := fixtures.TestFolderThreeFiles().(*factories.Folder)
	assert.Equal(t, true, ok)

	err := fixture.Write(".")
	assert.Equal(t, nil, err)

	files, err := fileSvc.ListFiles("./" + fixture.GetName())
	assert.Equal(t, nil, err)
	assert.Equal(t, []string{"SingleFile.txt", "SingleFile2.txt", "SingleFile3.txt"}, files)

	err = fixture.Remove(".")
	assert.Equal(t, nil, err)
}
