package downloading

import (
	"io"
	"net/http"
	"os"
)

type (
	// FileDownloader uses an interface to implement the DownloadFile() function.
	FileDownloader struct {
		// Dependencies points to an implementation of FileDownloaderDependencies.
		Dependencies FileDownloaderDependencies
	}

	// FileDownloaderRuntime is a runtime implementation for FileDownloaderDependencies.
	FileDownloaderRuntime struct {
	}
)

// DefaultDownloader returns a FileDownloader with runtime implementation.
func DefaultDownloader() FileDownloader {
	return FileDownloader{Dependencies: &FileDownloaderRuntime{}}
}

// NewHTTPRequest calls and returns http.NewRequest().
func (*FileDownloaderRuntime) NewHTTPRequest(
	method, url string,
	body io.Reader,
) (*http.Request, error) {
	return http.NewRequest(method, url, body)
}

// DoRequest calls and returns http.DefaultClient.Do().
func (*FileDownloaderRuntime) DoRequest(req *http.Request) (*http.Response, error) {
	return http.DefaultClient.Do(req)
}

// CreateFile calls and returns os.Create().
func (*FileDownloaderRuntime) CreateFile(relativePath string) (*os.File, error) {
	return os.Create(relativePath)
}
