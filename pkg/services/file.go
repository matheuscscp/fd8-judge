package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type (
	// FileUploadInfo holds HTTP fields for the upload request.
	FileUploadInfo struct {
		// Method is the method for the upload request.
		Method string `json:"method"`

		// URL is the URL for the upload request.
		URL string `json:"url"`

		// Headers holds the headers for the upload request.
		Headers http.Header `json:"headers"`
	}

	// FileService provides methods to manipulate files locally.
	FileService interface {
		// DownloadFile downloads a file and stores it in the local file system.
		DownloadFile(relativePath, url string, headers http.Header) (int64, error)

		// RequestUploadInfo requests an upload tuple (method, url, headers) to an authorized server.
		RequestUploadInfo(authorizedServerURL string, fileSize int) (*FileUploadInfo, error)

		// UploadFile uploads a file given an upload tuple (method, url, headers).
		UploadFile(relativePath string, uploadInfo *FileUploadInfo) error
	}

	// FileServiceRuntime is the contract to supply for the default implementation of FileService.
	FileServiceRuntime interface {
		// NewHTTPRequest returns a new *http.Request.
		NewHTTPRequest(method, url string, body io.Reader) (*http.Request, error)

		// DoRequest executes an *http.Request.
		DoRequest(req *http.Request) (*http.Response, error)

		// CreateFile creates a file in the given relativePath.
		CreateFile(relativePath string) (*os.File, error)

		// DoGetRequest executes an HTTP GET request.
		DoGetRequest(url string) (*http.Response, error)

		// DecodeUploadInfo reads a JSON representation of UploadInfo from a reader.
		DecodeUploadInfo(body io.Reader) (*FileUploadInfo, error)

		// OpenFile opens the file stored in relativePath and returns an io.ReadCloser.
		OpenFile(relativePath string) (io.ReadCloser, error)
	}

	// defaultFileService uses an interface to implement the DownloadFile() function.
	defaultFileService struct {
		runtime FileServiceRuntime
	}

	// fileServiceDefaultRuntime is a runtime implementation for FileServiceRuntime.
	fileServiceDefaultRuntime struct {
	}
)

// NewFileService returns a new instance of the default implementation of FileService.
// If nil is passed, the default FileService will be created with the default FileServiceRuntime.
func NewFileService(runtime FileServiceRuntime) FileService {
	if runtime == nil {
		runtime = &fileServiceDefaultRuntime{}
	}
	return &defaultFileService{runtime: runtime}
}

// DownloadFile downloads a file and stores it in the given relative path.
// The int64 return value is the number of bytes downloaded.
func (f *defaultFileService) DownloadFile(relativePath, url string, headers http.Header) (int64, error) {
	// create request object
	req, err := f.runtime.NewHTTPRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, &BuildFileDownloadRequestError{Wrapped: err}
	}
	for headerName, headerValues := range headers {
		for _, headerValue := range headerValues {
			req.Header.Add(headerName, headerValue)
		}
	}

	// do request
	resp, err := f.runtime.DoRequest(req)
	if err != nil {
		return 0, &DoFileDownloadRequestError{Wrapped: err}
	}
	defer resp.Body.Close()

	// check status
	if resp.StatusCode != http.StatusOK {
		return 0, &UnexpectedStatusInFileDownloadResponseError{Status: resp.Status}
	}

	// create file
	out, err := f.runtime.CreateFile(relativePath)
	if err != nil {
		return 0, &CreateFileForDownloadError{Wrapped: err}
	}
	defer out.Close()

	// download
	return io.Copy(out, resp.Body)
}

// RequestUploadInfo requests UploadInfo data to an authorized endpoint.
func (f *defaultFileService) RequestUploadInfo(authorizedServerURL string, fileSize int) (*FileUploadInfo, error) {
	// do request
	resp, err := f.runtime.DoGetRequest(fmt.Sprintf("%s?fileSize=%d", authorizedServerURL, fileSize))
	if err != nil {
		return nil, &RequestFileUploadInfoError{Wrapped: err}
	}
	defer resp.Body.Close()

	// check status
	if resp.StatusCode != http.StatusOK {
		return nil, &UnexpectedStatusInFileUploadInfoResponseError{Status: resp.Status}
	}

	// parse response
	uploadInfo, err := f.runtime.DecodeUploadInfo(resp.Body)
	if err != nil {
		return nil, &DecodeFileUploadInfoError{Wrapped: err}
	}

	return uploadInfo, nil
}

// UploadFile uploads a file stored in the given relative path.
func (f *defaultFileService) UploadFile(relativePath string, uploadInfo *FileUploadInfo) error {
	// open file
	file, err := f.runtime.OpenFile(relativePath)
	if err != nil {
		return &OpenUploadFileError{Wrapped: err}
	}
	defer file.Close()

	// create request object
	req, err := f.runtime.NewHTTPRequest(uploadInfo.Method, uploadInfo.URL, file)
	if err != nil {
		return &BuildFileUploadRequestError{Wrapped: err}
	}
	for headerName, headerValues := range uploadInfo.Headers {
		for _, headerValue := range headerValues {
			req.Header.Add(headerName, headerValue)
		}
	}

	// do request
	resp, err := f.runtime.DoRequest(req)
	if err != nil {
		return &DoFileUploadRequestError{Wrapped: err}
	}
	defer resp.Body.Close()

	// check status
	if resp.StatusCode != http.StatusOK {
		return &UnexpectedStatusInFileUploadResponseError{Status: resp.Status}
	}

	return nil
}

// NewHTTPRequest calls and returns http.NewRequest().
func (*fileServiceDefaultRuntime) NewHTTPRequest(
	method, url string,
	body io.Reader,
) (*http.Request, error) {
	return http.NewRequest(method, url, body)
}

// DoRequest calls and returns http.DefaultClient.Do().
func (*fileServiceDefaultRuntime) DoRequest(req *http.Request) (*http.Response, error) {
	return http.DefaultClient.Do(req)
}

// CreateFile calls and returns os.Create().
func (*fileServiceDefaultRuntime) CreateFile(relativePath string) (*os.File, error) {
	return os.Create(relativePath)
}

// DoGetRequest calls and returns http.Get().
func (*fileServiceDefaultRuntime) DoGetRequest(url string) (*http.Response, error) {
	return http.Get(url)
}

// DecodeUploadInfo wraps around json.NewDecoder().Decode().
func (*fileServiceDefaultRuntime) DecodeUploadInfo(reader io.Reader) (*FileUploadInfo, error) {
	var uploadInfo FileUploadInfo
	err := json.NewDecoder(reader).Decode(&uploadInfo)
	if err != nil {
		return nil, err
	}
	return &uploadInfo, nil
}

// OpenFile opens the file stored in relativePath and returns an io.ReadCloser.
func (*fileServiceDefaultRuntime) OpenFile(relativePath string) (io.ReadCloser, error) {
	return os.Open(relativePath)
}
