package uploading

import (
	"fmt"
	"net/http"

	"github.com/matheuscscp/fd8-judge/errors"
)

// RequestUploadInfo requests UploadInfo data to an authorized endpoint.
func (f *FileUploader) RequestUploadInfo(authorizedServerURL string, fileSize int) (*UploadInfo, error) {
	// do request
	resp, err := f.Dependencies.DoGetRequest(fmt.Sprintf("%s?fileSize=%d", authorizedServerURL, fileSize))
	if err != nil {
		return nil, &RequestUploadInfoError{WrapperError: errors.WrapperError{Wrapped: err}}
	}
	defer resp.Body.Close()

	// check status
	if resp.StatusCode != http.StatusOK {
		return nil, &UnexpectedStatusInUploadInfoResponseError{Status: resp.Status}
	}

	// parse response
	uploadInfo, err := f.Dependencies.DecodeUploadInfo(resp.Body)
	if err != nil {
		return nil, &DecodeUploadInfoError{WrapperError: errors.WrapperError{Wrapped: err}}
	}

	return uploadInfo, nil
}
