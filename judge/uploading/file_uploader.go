package uploading

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
)

type (
	// FileUploader uses an interface to implement the RequestUploadInfo() and UploadFile() functions.
	FileUploader struct {
		// Dependencies points to an implementation of FileUploaderDependencies.
		Dependencies FileUploaderDependencies
	}

	// FileUploaderRuntime is a runtime implementation for FileUploaderDependencies.
	FileUploaderRuntime struct {
	}
)

// DefaultUploader returns a FileUploader with runtime implementation.
func DefaultUploader() FileUploader {
	return FileUploader{Dependencies: &FileUploaderRuntime{}}
}

// DoGetRequest calls and returns http.Get().
func (*FileUploaderRuntime) DoGetRequest(url string) (*http.Response, error) {
	return http.Get(url)
}

// DecodeUploadInfo wraps around json.NewDecoder().Decode().
func (*FileUploaderRuntime) DecodeUploadInfo(reader io.Reader) (*UploadInfo, error) {
	var uploadInfo UploadInfo
	err := json.NewDecoder(reader).Decode(&uploadInfo)
	if err != nil {
		return nil, err
	}
	return &uploadInfo, nil
}

// OpenFile opens the file stored in relativePath and returns an io.ReadCloser.
func (*FileUploaderRuntime) OpenFile(relativePath string) (io.ReadCloser, error) {
	return os.Open(relativePath)
}

// NewHTTPRequest returns a new *http.Request.
func (*FileUploaderRuntime) NewHTTPRequest(method, url string, body io.Reader) (*http.Request, error) {
	return http.NewRequest(method, url, body)
}

// DoRequest calls and returns http.DefaultClient.Do().
func (*FileUploaderRuntime) DoRequest(req *http.Request) (*http.Response, error) {
	return http.DefaultClient.Do(req)
}
