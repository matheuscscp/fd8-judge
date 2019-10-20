package downloading

import (
	"io"
	"net/http"
)

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
