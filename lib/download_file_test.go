// +build integration

package lib_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"testing"

	"github.com/matheuscscp/fd8-judge/lib"
	"github.com/matheuscscp/fd8-judge/testing/factory"
)

func TestDownloadFile(t *testing.T) {
	// create server
	f := factory.NewHTTPServerFactory()
	listener, server, err := f.NewDummy()
	if err != nil {
		t.Fatalf("error creating HTTP server: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port

	// download bytes
	const relativePath = "./TestDownloadFile.tmp"
	const payload = "PAYLOAD"
	const bytesToBeDownloaded = int64(len(payload))
	bytesDownloaded, err := lib.DownloadFile(
		relativePath,
		fmt.Sprintf("http://localhost:%d/dummy", port),
		nil,
	)
	if err != nil {
		t.Fatalf("error downloading file: %v", err)
	}
	if bytesDownloaded != bytesToBeDownloaded {
		t.Fatalf(
			"wrong number of bytes downloaded, want %d, got %d",
			bytesToBeDownloaded,
			bytesDownloaded,
		)
	}

	// check file content
	fileContentBytes, err := ioutil.ReadFile(relativePath)
	if err != nil {
		t.Fatalf("error reading downloaded file: %v", err)
	}
	fileContent := string(fileContentBytes)
	if fileContent != payload {
		t.Fatalf("wrong downloaded file, want '%s', got '%s'", payload, fileContent)
	}

	// erase file
	err = os.Remove(relativePath)
	if err != nil {
		t.Fatalf("error removing downloaded file: %v", err)
	}

	// shutdown test server
	err = server.Shutdown(context.Background())
	if err != nil {
		t.Fatalf("error shutting down test server: %v", err)
	}
}
