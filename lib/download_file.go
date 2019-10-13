package lib

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

// DownloadFile downloads a file and stores it in the given relative path.
// The int64 return value is the number of bytes downloaded.
func DownloadFile(relativePath, url string, headers http.Header) (int64, error) {
	// create request object
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to build download request: %v", err)
	}
	for headerName, headerValues := range headers {
		for _, headerValue := range headerValues {
			req.Header.Add(headerName, headerValue)
		}
	}

	// do request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	// check status
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("error status code in download response: %s", resp.Status)
	}

	// create file
	out, err := os.Create(relativePath)
	if err != nil {
		return 0, err
	}
	defer out.Close()

	// download
	return io.Copy(out, resp.Body)
}
