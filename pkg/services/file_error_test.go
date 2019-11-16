// +build unit

package services_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/matheuscscp/fd8-judge/pkg/services"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestDownloadFileError(t *testing.T) {
	var mockRuntime *services.MockFileServiceRuntime

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
			errStr    string
			errUnwrap error
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
				err: &services.BuildFileDownloadRequestError{Wrapped: fmt.Errorf("error")},
			},
			outProps: testOutputProps{
				errStr:    "error building download request: error",
				errUnwrap: fmt.Errorf("error"),
			},
			mocks: func() {
				mockRuntime.EXPECT().NewHTTPRequest(http.MethodGet, "", nil).Return(nil, fmt.Errorf("error"))
			},
		},
		"do-download-request-error": {
			output: testOutput{
				err: &services.DoFileDownloadRequestError{Wrapped: fmt.Errorf("error")},
			},
			outProps: testOutputProps{
				errStr:    "error performing download request: error",
				errUnwrap: fmt.Errorf("error"),
			},
			mocks: func() {
				mockRuntime.EXPECT().NewHTTPRequest(http.MethodGet, "", nil).Return(nil, nil)
				mockRuntime.EXPECT().DoRequest(nil).Return(nil, fmt.Errorf("error"))
			},
		},
		"unexpected-status-in-download-response-error": {
			output: testOutput{
				err: &services.UnexpectedStatusInFileDownloadResponseError{Status: "status"},
			},
			outProps: testOutputProps{
				errStr: "unexpected status in download response: status",
			},
			mocks: func() {
				mockRuntime.EXPECT().NewHTTPRequest(http.MethodGet, "", nil).Return(nil, nil)
				mockRuntime.EXPECT().DoRequest(nil).Return(&http.Response{
					StatusCode: 201,
					Status:     "status",
					Body:       ioutil.NopCloser(nil),
				}, nil)
			},
		},
		"create-file-error": {
			output: testOutput{
				err: &services.CreateFileForDownloadError{Wrapped: fmt.Errorf("error")},
			},
			outProps: testOutputProps{
				errStr:    "error creating file for downloaded data: error",
				errUnwrap: fmt.Errorf("error"),
			},
			mocks: func() {
				mockRuntime.EXPECT().NewHTTPRequest(http.MethodGet, "", nil).Return(nil, nil)
				mockRuntime.EXPECT().DoRequest(nil).Return(&http.Response{
					StatusCode: 200,
					Body:       ioutil.NopCloser(nil),
				}, nil)
				mockRuntime.EXPECT().CreateFile("").Return(nil, fmt.Errorf("error"))
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// mocks
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRuntime = services.NewMockFileServiceRuntime(ctrl)
			test.mocks()

			fileSvc := services.NewFileService(mockRuntime)
			bytes, err := fileSvc.DownloadFile(test.input.relativePath, test.input.url, test.input.headers)
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

func TestRequestUploadInfoError(t *testing.T) {
	var mockRuntime *services.MockFileServiceRuntime

	type (
		testInput struct {
			authorizedServerURL string
			fileSize            int
		}
		testOutput struct {
			uploadInfo *services.FileUploadInfo
			err        error
		}
		testOutputProps struct {
			errStr    string
			errUnwrap error
		}
	)
	var tests = map[string]struct {
		input    testInput
		output   testOutput
		outProps testOutputProps
		mocks    func()
	}{
		"request-upload-info-error": {
			output: testOutput{
				err: &services.RequestFileUploadInfoError{Wrapped: fmt.Errorf("error")},
			},
			outProps: testOutputProps{
				errStr:    "error requesting upload info: error",
				errUnwrap: fmt.Errorf("error"),
			},
			mocks: func() {
				mockRuntime.EXPECT().DoGetRequest("?fileSize=0").Return(nil, fmt.Errorf("error"))
			},
		},
		"unexpected-status-in-download-response-error": {
			output: testOutput{
				err: &services.UnexpectedStatusInFileUploadInfoResponseError{Status: "status"},
			},
			outProps: testOutputProps{
				errStr: "unexpected status in upload info response: status",
			},
			mocks: func() {
				mockRuntime.EXPECT().DoGetRequest("?fileSize=0").Return(&http.Response{
					StatusCode: 201,
					Status:     "status",
					Body:       ioutil.NopCloser(nil),
				}, nil)
			},
		},
		"decode-upload-info-error": {
			output: testOutput{
				err: &services.DecodeFileUploadInfoError{Wrapped: fmt.Errorf("error")},
			},
			outProps: testOutputProps{
				errStr:    "error decoding upload info: error",
				errUnwrap: fmt.Errorf("error"),
			},
			mocks: func() {
				mockRuntime.EXPECT().DoGetRequest("?fileSize=0").Return(&http.Response{
					StatusCode: 200,
					Status:     "status",
					Body:       ioutil.NopCloser(nil),
				}, nil)
				mockRuntime.EXPECT().DecodeUploadInfo(ioutil.NopCloser(nil)).Return(nil, fmt.Errorf("error"))
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// mocks
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRuntime = services.NewMockFileServiceRuntime(ctrl)
			test.mocks()

			fileSvc := services.NewFileService(mockRuntime)
			uploadInfo, err := fileSvc.RequestUploadInfo(test.input.authorizedServerURL, test.input.fileSize)
			errStr := ""
			if err != nil {
				errStr = err.Error()
			}
			assert.Equal(t, test.output.uploadInfo, uploadInfo)
			assert.Equal(t, test.output.err, err)
			assert.Equal(t, test.outProps.errStr, errStr)
		})
	}
}

func TestUploadFileError(t *testing.T) {
	var mockRuntime *services.MockFileServiceRuntime

	type (
		testInput struct {
			relativePath string
			uploadInfo   *services.FileUploadInfo
		}
		testOutput struct {
			err error
		}
		testOutputProps struct {
			errStr    string
			errUnwrap error
		}
	)
	var tests = map[string]struct {
		input    testInput
		output   testOutput
		outProps testOutputProps
		mocks    func()
	}{
		"open-upload-file-error": {
			output: testOutput{
				err: &services.OpenUploadFileError{Wrapped: fmt.Errorf("error")},
			},
			outProps: testOutputProps{
				errStr:    "error opening upload file: error",
				errUnwrap: fmt.Errorf("error"),
			},
			mocks: func() {
				mockRuntime.EXPECT().OpenFile("").Return(nil, fmt.Errorf("error"))
			},
		},
		"build-upload-request-error": {
			input: testInput{
				uploadInfo: &services.FileUploadInfo{},
			},
			output: testOutput{
				err: &services.BuildFileUploadRequestError{Wrapped: fmt.Errorf("error")},
			},
			outProps: testOutputProps{
				errStr:    "error building upload request: error",
				errUnwrap: fmt.Errorf("error"),
			},
			mocks: func() {
				mockRuntime.EXPECT().OpenFile("").Return(ioutil.NopCloser(nil), nil)
				mockRuntime.EXPECT().NewHTTPRequest("", "", ioutil.NopCloser(nil)).Return(nil, fmt.Errorf("error"))
			},
		},
		"do-upload-request-error": {
			input: testInput{
				uploadInfo: &services.FileUploadInfo{},
			},
			output: testOutput{
				err: &services.DoFileUploadRequestError{Wrapped: fmt.Errorf("error")},
			},
			outProps: testOutputProps{
				errStr:    "error performing upload request: error",
				errUnwrap: fmt.Errorf("error"),
			},
			mocks: func() {
				mockRuntime.EXPECT().OpenFile("").Return(ioutil.NopCloser(nil), nil)
				mockRuntime.EXPECT().NewHTTPRequest("", "", ioutil.NopCloser(nil)).Return(&http.Request{}, nil)
				mockRuntime.EXPECT().DoRequest(&http.Request{}).Return(nil, fmt.Errorf("error"))
			},
		},
		"unexpected-status-in-upload-response-error": {
			input: testInput{
				uploadInfo: &services.FileUploadInfo{},
			},
			output: testOutput{
				err: &services.UnexpectedStatusInFileUploadResponseError{Status: "status"},
			},
			outProps: testOutputProps{
				errStr: "unexpected status in upload response: status",
			},
			mocks: func() {
				mockRuntime.EXPECT().OpenFile("").Return(ioutil.NopCloser(nil), nil)
				mockRuntime.EXPECT().NewHTTPRequest("", "", ioutil.NopCloser(nil)).Return(&http.Request{}, nil)
				mockRuntime.EXPECT().DoRequest(&http.Request{}).Return(&http.Response{
					StatusCode: 201,
					Status:     "status",
					Body:       ioutil.NopCloser(nil),
				}, nil)
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// mocks
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRuntime = services.NewMockFileServiceRuntime(ctrl)
			test.mocks()

			fileSvc := services.NewFileService(mockRuntime)
			err := fileSvc.UploadFile(test.input.relativePath, test.input.uploadInfo)
			errStr := ""
			if err != nil {
				errStr = err.Error()
			}
			assert.Equal(t, test.output.err, err)
			assert.Equal(t, test.outProps.errStr, errStr)
		})
	}
}
