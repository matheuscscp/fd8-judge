// +build unit

package uploading_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/matheuscscp/fd8-judge/errors"
	"github.com/matheuscscp/fd8-judge/judge/uploading"
	"github.com/stretchr/testify/assert"
)

func TestUploadFileError(t *testing.T) {
	var mockDependencies *uploading.MockFileUploaderDependencies

	type (
		testInput struct {
			relativePath string
			uploadInfo   *uploading.UploadInfo
		}
		testOutput struct {
			err error
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
		"open-upload-file-error": {
			output: testOutput{
				err: &uploading.OpenUploadFileError{WrapperError: errors.WrapperError{Wrapped: fmt.Errorf("error")}},
			},
			outProps: testOutputProps{
				errStr: "error opening upload file: error",
			},
			mocks: func() {
				mockDependencies.EXPECT().OpenFile("").Return(nil, fmt.Errorf("error"))
			},
		},
		"build-upload-request-error": {
			input: testInput{
				uploadInfo: &uploading.UploadInfo{},
			},
			output: testOutput{
				err: &uploading.BuildUploadRequestError{WrapperError: errors.WrapperError{Wrapped: fmt.Errorf("error")}},
			},
			outProps: testOutputProps{
				errStr: "error building upload request: error",
			},
			mocks: func() {
				mockDependencies.EXPECT().OpenFile("").Return(ioutil.NopCloser(nil), nil)
				mockDependencies.EXPECT().NewHTTPRequest("", "", ioutil.NopCloser(nil)).Return(nil, fmt.Errorf("error"))
			},
		},
		"do-upload-request-error": {
			input: testInput{
				uploadInfo: &uploading.UploadInfo{},
			},
			output: testOutput{
				err: &uploading.DoUploadRequestError{WrapperError: errors.WrapperError{Wrapped: fmt.Errorf("error")}},
			},
			outProps: testOutputProps{
				errStr: "error performing upload request: error",
			},
			mocks: func() {
				mockDependencies.EXPECT().OpenFile("").Return(ioutil.NopCloser(nil), nil)
				mockDependencies.EXPECT().NewHTTPRequest("", "", ioutil.NopCloser(nil)).Return(&http.Request{}, nil)
				mockDependencies.EXPECT().DoRequest(&http.Request{}).Return(nil, fmt.Errorf("error"))
			},
		},
		"unexpected-status-in-upload-response-error": {
			input: testInput{
				uploadInfo: &uploading.UploadInfo{},
			},
			output: testOutput{
				err: &uploading.UnexpectedStatusInUploadResponseError{Status: "status"},
			},
			outProps: testOutputProps{
				errStr: "unexpected status in upload response: status",
			},
			mocks: func() {
				mockDependencies.EXPECT().OpenFile("").Return(ioutil.NopCloser(nil), nil)
				mockDependencies.EXPECT().NewHTTPRequest("", "", ioutil.NopCloser(nil)).Return(&http.Request{}, nil)
				mockDependencies.EXPECT().DoRequest(&http.Request{}).Return(&http.Response{
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
			mockDependencies = uploading.NewMockFileUploaderDependencies(ctrl)
			test.mocks()

			uploader := uploading.FileUploader{Dependencies: mockDependencies}
			err := uploader.UploadFile(test.input.relativePath, test.input.uploadInfo)
			errStr := ""
			if err != nil {
				errStr = err.Error()
			}
			assert.Equal(t, test.output.err, err)
			assert.Equal(t, test.outProps.errStr, errStr)
		})
	}
}
