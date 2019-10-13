package downloading

import (
	"io"
	"net/http"
	"os"
)

// FileDownloaderDependencies is the contract to supply for FileDownloader.
type FileDownloaderDependencies interface {
	// NewHTTPRequest returns a new *http.Request.
	NewHTTPRequest(string, string, io.Reader) (*http.Request, error)

	// DoRequest executes an *http.Request.
	DoRequest(*http.Request) (*http.Response, error)

	// CreateFile creates a file.
	CreateFile(string) (*os.File, error)
}
