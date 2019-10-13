// +build unit

package downloading_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/matheuscscp/fd8-judge/judge/downloading"
	"github.com/stretchr/testify/assert"
)

func TestDownloadFileError(t *testing.T) {
	var mockDependencies *downloading.MockFileDownloaderDependencies

	var tests = map[string]struct {
		relativePath string
		url          string
		headers      http.Header
		bytes        int64
		err          error
		errStr       string
		errUnwrap    error
		mocks        func()
	}{
		"build-download-request-error": {
			url:       "mockURL",
			err:       &downloading.BuildDownloadRequestError{Wrapped: fmt.Errorf("error")},
			errStr:    "failed to build download request: error",
			errUnwrap: fmt.Errorf("error"),
			mocks: func() {
				mockDependencies.EXPECT().NewHTTPRequest(http.MethodGet, "mockURL", nil).Return(nil, fmt.Errorf("error"))
			},
		},
		"do-download-request-error": {
			url:       "mockURL",
			err:       &downloading.DoDownloadRequestError{Wrapped: fmt.Errorf("error")},
			errStr:    "failed to do download request: error",
			errUnwrap: fmt.Errorf("error"),
			mocks: func() {
				mockDependencies.EXPECT().NewHTTPRequest(http.MethodGet, "mockURL", nil).Return(nil, nil)
				mockDependencies.EXPECT().DoRequest(nil).Return(nil, fmt.Errorf("error"))
			},
		},
		"unexpected-status-in-download-response-error": {
			url:    "mockURL",
			err:    &downloading.UnexpectedStatusInDownloadResponseError{Status: "status"},
			errStr: "received unexpected status in download response: status",
			mocks: func() {
				mockDependencies.EXPECT().NewHTTPRequest(http.MethodGet, "mockURL", nil).Return(nil, nil)
				mockDependencies.EXPECT().DoRequest(nil).Return(&http.Response{
					StatusCode: 201,
					Status:     "status",
					Body:       ioutil.NopCloser(nil),
				}, nil)
			},
		},
		"create-file-error": {
			url:       "mockURL",
			err:       &downloading.CreateFileError{Wrapped: fmt.Errorf("error")},
			errStr:    "failed to create file to store downloaded data: error",
			errUnwrap: fmt.Errorf("error"),
			mocks: func() {
				mockDependencies.EXPECT().NewHTTPRequest(http.MethodGet, "mockURL", nil).Return(nil, nil)
				mockDependencies.EXPECT().DoRequest(nil).Return(&http.Response{
					StatusCode: 200,
					Body:       ioutil.NopCloser(nil),
				}, nil)
				mockDependencies.EXPECT().CreateFile("").Return(nil, fmt.Errorf("error"))
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// mock
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockDependencies = downloading.NewMockFileDownloaderDependencies(ctrl)
			test.mocks()

			downloader := downloading.FileDownloader{Dependencies: mockDependencies}
			bytes, err := downloader.DownloadFile(test.relativePath, test.url, test.headers)
			errStr := ""
			if err != nil {
				errStr = err.Error()
			}
			errUnwrap := errors.Unwrap(err)
			assert.Equal(t, test.bytes, bytes)
			assert.Equal(t, test.err, err)
			assert.Equal(t, test.errStr, errStr)
			assert.Equal(t, test.errUnwrap, errUnwrap)
		})
	}
}
