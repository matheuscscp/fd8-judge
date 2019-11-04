// +build integration

package uploading_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"

	"github.com/matheuscscp/fd8-judge/judge/uploading"
	"github.com/matheuscscp/fd8-judge/testing/factory"
	"github.com/stretchr/testify/assert"
)

func TestRequestUploadInfo(t *testing.T) {
	// create server
	f := factory.NewHTTPServerFactory()
	listener, server, err := f.NewDummyUploader()
	assert.Equal(t, nil, err)
	port := listener.Addr().(*net.TCPAddr).Port

	// request upload info
	uploader := uploading.DefaultUploader()
	uploadInfo, err := uploader.RequestUploadInfo(
		fmt.Sprintf("http://localhost:%d/upload-info", port),
		5,
	)
	assert.Equal(t, nil, err)
	assert.Equal(t, &uploading.UploadInfo{
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
