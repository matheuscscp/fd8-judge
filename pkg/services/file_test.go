// +build integration

package services_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"testing"

	"github.com/matheuscscp/fd8-judge/pkg/services"
	"github.com/matheuscscp/fd8-judge/test/factories"
	"github.com/matheuscscp/fd8-judge/test/fixtures"
	"github.com/stretchr/testify/assert"
)

func TestDownloadFile(t *testing.T) {
	// create server
	f := factories.NewHTTPServerFactory()
	listener, server, err := f.NewDummy()
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

func TestRequestUploadInfo(t *testing.T) {
	// create server
	f := factories.NewHTTPServerFactory()
	listener, server, err := f.NewDummyUploader()
	assert.Equal(t, nil, err)
	port := listener.Addr().(*net.TCPAddr).Port

	// request upload info
	fileSvc := services.NewFileService(nil)
	uploadInfo, err := fileSvc.RequestUploadInfo(
		fmt.Sprintf("http://localhost:%d/upload-info", port),
		5,
	)
	assert.Equal(t, nil, err)
	assert.Equal(t, &services.FileUploadInfo{
		Method: http.MethodPut,
		URL:    fmt.Sprintf("http://localhost:%d/upload", port),
		Headers: http.Header{
			"Content-Length": []string{"5"},
		},
	}, uploadInfo)

	// shutdown test server
	err = server.Shutdown(context.Background())
	assert.Equal(t, nil, err)
}

func TestUploadFile(t *testing.T) {
	// create server
	f := factories.NewHTTPServerFactory()
	listener, server, err := f.NewDummyUploader()
	assert.Equal(t, nil, err)
	port := listener.Addr().(*net.TCPAddr).Port
	url := fmt.Sprintf("http://localhost:%d/upload", port)

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
	err = fileSvc.UploadFile(relativePath, &services.FileUploadInfo{
		Method: http.MethodPut,
		URL:    url,
		Headers: http.Header{
			"Content-Length": []string{fmt.Sprintf("%d", bytesToBeUploaded)},
		},
	})
	assert.Equal(t, nil, err)

	// check uploaded content
	resp, err := http.Get(url)
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
		uncompressedRelativePath string
	}{
		"single-file": {
			fixture:                  fixtures.SingleFile(),
			expectedFileTree:         fixtures.SingleFile(),
			inputRelativePath:        "./SingleFile.txt",
			uncompressedRelativePath: "./SingleFile.txt",
		},
		"empty-folder": {
			fixture:                  fixtures.EmptyFolder(),
			expectedFileTree:         fixtures.EmptyFolder(),
			inputRelativePath:        "./EmptyFolder",
			uncompressedRelativePath: "./EmptyFolder",
		},
		"big-folder": {
			fixture:                  fixtures.TestFolder(),
			expectedFileTree:         fixtures.TestFolder(),
			inputRelativePath:        "./TestFolder",
			uncompressedRelativePath: "./TestFolder",
		},
		"cut-path-before-single-file": {
			fixture:                  fixtures.TestFolderOneFile(),
			expectedFileTree:         fixtures.SingleFile(),
			inputRelativePath:        "./TestFolderOneFile/SingleFile.txt",
			uncompressedRelativePath: "./SingleFile.txt",
		},
		"cut-path-before-empty-folder": {
			fixture:                  fixtures.TestFolderOneFolder(),
			expectedFileTree:         fixtures.EmptyFolder(),
			inputRelativePath:        "./TestFolderOneFolder/EmptyFolder",
			uncompressedRelativePath: "./EmptyFolder",
		},
		"cut-long-path-before-big-folder": {
			fixture:                  fixtures.TestDummyRootFolder(),
			expectedFileTree:         fixtures.TestFolder(),
			inputRelativePath:        "./TestDummyRootFolder//MiddleFolder/TestFolder",
			uncompressedRelativePath: "./TestFolder",
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

			err = fileSvc.Uncompress("./TestCompressedFile.tar.gz", ".")
			assert.Equal(t, nil, err)

			err = os.Remove("./TestCompressedFile.tar.gz")
			assert.Equal(t, nil, err)

			fileTree, err := factories.ReadFileTree(test.uncompressedRelativePath, true)
			assert.Equal(t, nil, err)
			assert.Equal(t, test.expectedFileTree, fileTree)

			err = fileTree.Remove(".")
			assert.Equal(t, nil, err)
		})
	}
}
