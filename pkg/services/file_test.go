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
	"github.com/matheuscscp/fd8-judge/testing/factory"
	"github.com/stretchr/testify/assert"
)

func TestDownloadFile(t *testing.T) {
	// create server
	f := factory.NewHTTPServerFactory()
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
	f := factory.NewHTTPServerFactory()
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
	f := factory.NewHTTPServerFactory()
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
	err = ioutil.WriteFile(relativePath, []byte(payload), 0644)
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
