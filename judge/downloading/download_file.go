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

// DownloadFile downloads a file and stores it in the given relative path.
// The int64 return value is the number of bytes downloaded.
func (f *FileDownloader) DownloadFile(relativePath, url string, headers http.Header) (int64, error) {
	// create request object
	req, err := f.Dependencies.NewHTTPRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, &BuildDownloadRequestError{Wrapped: err}
	}
	for headerName, headerValues := range headers {
		for _, headerValue := range headerValues {
			req.Header.Add(headerName, headerValue)
		}
	}

	// do request
	resp, err := f.Dependencies.DoRequest(req)
	if err != nil {
		return 0, &DoDownloadRequestError{Wrapped: err}
	}
	defer resp.Body.Close()

	// check status
	if resp.StatusCode != http.StatusOK {
		return 0, &UnexpectedStatusInDownloadResponseError{Status: resp.Status}
	}

	// create file
	out, err := f.Dependencies.CreateFile(relativePath)
	if err != nil {
		return 0, &CreateFileError{Wrapped: err}
	}
	defer out.Close()

	// download
	return io.Copy(out, resp.Body)
}

// NewHTTPRequest calls and returns http.NewRequest().
func (i *FileDownloaderRuntime) NewHTTPRequest(
	method, url string,
	body io.Reader,
) (*http.Request, error) {
	return http.NewRequest(method, url, body)
}

// DoRequest calls and returns http.DefaultClient.Do().
func (i *FileDownloaderRuntime) DoRequest(req *http.Request) (*http.Response, error) {
	return http.DefaultClient.Do(req)
}

// CreateFile calls and returns os.Create().
func (i *FileDownloaderRuntime) CreateFile(relativePath string) (*os.File, error) {
	return os.Create(relativePath)
}
