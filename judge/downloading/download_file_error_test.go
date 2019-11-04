// +build unit

package downloading_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/matheuscscp/fd8-judge/errors"
	"github.com/matheuscscp/fd8-judge/judge/downloading"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestDownloadFileError(t *testing.T) {
	var mockDependencies *downloading.MockFileDownloaderDependencies

	type (
		testInput struct {
			relativePath string
			url          string
			headers      http.Header
		}
		testOutput struct {
			bytes int64
			err   error
		}
		testOutputProps struct {
			errStr string
		}
	)
	var tests = map[string]struct {
		input    testInput
		output   testOutput
		outProps testOutputProps
		mocks    func()
	}{
		"build-download-request-error": {
			output: testOutput{
				err: &downloading.BuildDownloadRequestError{WrapperError: errors.WrapperError{Wrapped: fmt.Errorf("error")}},
			},
			outProps: testOutputProps{
				errStr: "error building download request: error",
			},
			mocks: func() {
				mockDependencies.EXPECT().NewHTTPRequest(http.MethodGet, "", nil).Return(nil, fmt.Errorf("error"))
			},
		},
		"do-download-request-error": {
			output: testOutput{
				err: &downloading.DoDownloadRequestError{WrapperError: errors.WrapperError{Wrapped: fmt.Errorf("error")}},
			},
			outProps: testOutputProps{
				errStr: "error performing download request: error",
			},
			mocks: func() {
				mockDependencies.EXPECT().NewHTTPRequest(http.MethodGet, "", nil).Return(nil, nil)
				mockDependencies.EXPECT().DoRequest(nil).Return(nil, fmt.Errorf("error"))
			},
		},
		"unexpected-status-in-download-response-error": {
			output: testOutput{
				err: &downloading.UnexpectedStatusInDownloadResponseError{Status: "status"},
			},
			outProps: testOutputProps{
				errStr: "unexpected status in download response: status",
			},
			mocks: func() {
				mockDependencies.EXPECT().NewHTTPRequest(http.MethodGet, "", nil).Return(nil, nil)
				mockDependencies.EXPECT().DoRequest(nil).Return(&http.Response{
					StatusCode: 201,
					Status:     "status",
					Body:       ioutil.NopCloser(nil),
				}, nil)
			},
		},
		"create-file-error": {
			output: testOutput{
				err: &downloading.CreateFileError{WrapperError: errors.WrapperError{Wrapped: fmt.Errorf("error")}},
			},
			outProps: testOutputProps{
				errStr: "error creating file for downloaded data: error",
			},
			mocks: func() {
				mockDependencies.EXPECT().NewHTTPRequest(http.MethodGet, "", nil).Return(nil, nil)
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
			// mocks
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockDependencies = downloading.NewMockFileDownloaderDependencies(ctrl)
			test.mocks()

			downloader := downloading.FileDownloader{Dependencies: mockDependencies}
			bytes, err := downloader.DownloadFile(test.input.relativePath, test.input.url, test.input.headers)
			errStr := ""
			if err != nil {
				errStr = err.Error()
			}
			assert.Equal(t, test.output.bytes, bytes)
			assert.Equal(t, test.output.err, err)
			assert.Equal(t, test.outProps.errStr, errStr)
		})
	}
}
