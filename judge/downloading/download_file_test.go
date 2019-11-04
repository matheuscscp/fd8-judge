// +build integration

package downloading_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"testing"

	"github.com/matheuscscp/fd8-judge/judge/downloading"
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
	downloader := downloading.DefaultDownloader()
	bytesDownloaded, err := downloader.DownloadFile(
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
