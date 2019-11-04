package uploading

import (
	"io"
	"net/http"
)

type (
	// UploadInfo holds HTTP fields for the upload request.
	UploadInfo struct {
		// Method is the method for the upload request.
		Method string `json:"method"`

		// URL is the URL for the upload request.
		URL string `json:"url"`

		// Headers holds the headers for the upload request.
		Headers http.Header `json:"headers"`
	}

	// FileUploaderDependencies is the contract to supply for FileUploader.
	FileUploaderDependencies interface {
		// DoGetRequest executes an HTTP GET request.
		DoGetRequest(url string) (*http.Response, error)

		// DecodeUploadInfo reads a JSON representation of UploadInfo from a reader.
		DecodeUploadInfo(body io.Reader) (*UploadInfo, error)

		// OpenFile opens the file stored in relativePath and returns an io.ReadCloser.
		OpenFile(relativePath string) (io.ReadCloser, error)

		// NewHTTPRequest returns a new *http.Request.
		NewHTTPRequest(method, url string, body io.Reader) (*http.Request, error)

		// DoRequest executes an *http.Request.
		DoRequest(req *http.Request) (*http.Response, error)
	}
)
