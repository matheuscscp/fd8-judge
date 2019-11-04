package downloading

import (
	"io"
	"net/http"
	"os"
)

// FileDownloaderDependencies is the contract to supply for FileDownloader.
type FileDownloaderDependencies interface {
	// NewHTTPRequest returns a new *http.Request.
	NewHTTPRequest(method, url string, body io.Reader) (*http.Request, error)

	// DoRequest executes an *http.Request.
	DoRequest(req *http.Request) (*http.Response, error)

	// CreateFile creates a file in the given relativePath.
	CreateFile(relativePath string) (*os.File, error)
}
