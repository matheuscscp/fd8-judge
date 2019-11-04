// +build unit

package uploading_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/matheuscscp/fd8-judge/errors"
	"github.com/matheuscscp/fd8-judge/judge/uploading"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestRequestUploadInfoError(t *testing.T) {
	var mockDependencies *uploading.MockFileUploaderDependencies

	type (
		testInput struct {
			authorizedServerURL string
			fileSize            int
		}
		testOutput struct {
			uploadInfo *uploading.UploadInfo
			err        error
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
		"request-upload-info-error": {
			output: testOutput{
				err: &uploading.RequestUploadInfoError{WrapperError: errors.WrapperError{Wrapped: fmt.Errorf("error")}},
			},
			outProps: testOutputProps{
				errStr: "error requesting upload info: error",
			},
			mocks: func() {
				mockDependencies.EXPECT().DoGetRequest("?fileSize=0").Return(nil, fmt.Errorf("error"))
			},
		},
		"unexpected-status-in-download-response-error": {
			output: testOutput{
				err: &uploading.UnexpectedStatusInUploadInfoResponseError{Status: "status"},
			},
			outProps: testOutputProps{
				errStr: "unexpected status in upload info response: status",
			},
			mocks: func() {
				mockDependencies.EXPECT().DoGetRequest("?fileSize=0").Return(&http.Response{
					StatusCode: 201,
					Status:     "status",
					Body:       ioutil.NopCloser(nil),
				}, nil)
			},
		},
		"decode-upload-info-error": {
			output: testOutput{
				err: &uploading.DecodeUploadInfoError{WrapperError: errors.WrapperError{Wrapped: fmt.Errorf("error")}},
			},
			outProps: testOutputProps{
				errStr: "error decoding upload info: error",
			},
			mocks: func() {
				mockDependencies.EXPECT().DoGetRequest("?fileSize=0").Return(&http.Response{
					StatusCode: 200,
					Status:     "status",
					Body:       ioutil.NopCloser(nil),
				}, nil)
				mockDependencies.EXPECT().DecodeUploadInfo(ioutil.NopCloser(nil)).Return(nil, fmt.Errorf("error"))
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
			uploadInfo, err := uploader.RequestUploadInfo(test.input.authorizedServerURL, test.input.fileSize)
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
