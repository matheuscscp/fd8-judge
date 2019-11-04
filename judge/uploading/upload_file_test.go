// +build integration

package uploading_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"testing"

	"github.com/matheuscscp/fd8-judge/judge/uploading"
	"github.com/matheuscscp/fd8-judge/testing/factory"
	"github.com/stretchr/testify/assert"
)

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
	uploader := uploading.DefaultUploader()
	err = uploader.UploadFile(relativePath, &uploading.UploadInfo{
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
