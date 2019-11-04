package uploading

import (
	"net/http"

	"github.com/matheuscscp/fd8-judge/errors"
)

// UploadFile uploads a file stored in the given relative path.
func (f *FileUploader) UploadFile(relativePath string, uploadInfo *UploadInfo) error {
	// open file
	file, err := f.Dependencies.OpenFile(relativePath)
	if err != nil {
		return &OpenUploadFileError{WrapperError: errors.WrapperError{Wrapped: err}}
	}
	defer file.Close()

	// create request object
	req, err := f.Dependencies.NewHTTPRequest(uploadInfo.Method, uploadInfo.URL, file)
	if err != nil {
		return &BuildUploadRequestError{WrapperError: errors.WrapperError{Wrapped: err}}
	}
	for headerName, headerValues := range uploadInfo.Headers {
		for _, headerValue := range headerValues {
			req.Header.Add(headerName, headerValue)
		}
	}

	// do request
	resp, err := f.Dependencies.DoRequest(req)
	if err != nil {
		return &DoUploadRequestError{WrapperError: errors.WrapperError{Wrapped: err}}
	}
	defer resp.Body.Close()

	// check status
	if resp.StatusCode != http.StatusOK {
		return &UnexpectedStatusInUploadResponseError{Status: resp.Status}
	}

	return nil
}
