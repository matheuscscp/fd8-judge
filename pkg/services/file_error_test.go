// +build unit

package services_test

import (
	"archive/tar"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/matheuscscp/fd8-judge/pkg/services"
	"github.com/matheuscscp/fd8-judge/testing/fixtures"
	"github.com/matheuscscp/fd8-judge/testing/mocks"
	mockServices "github.com/matheuscscp/fd8-judge/testing/mocks/pkg/services"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestDownloadFileError(t *testing.T) {
	var mockRuntime *mockServices.MockFileServiceRuntime

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
					Body:       &fixtures.NopReadCloser{},
				}, nil)
			},
		},
		"create-file-error": {
			output: testOutput{
				err: &services.CreateFileForDownloadError{Wrapped: fmt.Errorf("error")},
			},
			outProps: testOutputProps{
				errStr:    "error creating file for download data: error",
				errUnwrap: fmt.Errorf("error"),
			},
			mocks: func() {
				mockRuntime.EXPECT().NewHTTPRequest(http.MethodGet, "", nil).Return(nil, nil)
				mockRuntime.EXPECT().DoRequest(nil).Return(&http.Response{
					StatusCode: 200,
					Body:       &fixtures.NopReadCloser{},
				}, nil)
				mockRuntime.EXPECT().CreateFile("").Return(nil, fmt.Errorf("error"))
			},
		},
		"transfer-and-store-download-data-error": {
			output: testOutput{
				err: &services.TransferAndStoreDownloadFileError{Wrapped: fmt.Errorf("error")},
			},
			outProps: testOutputProps{
				errStr:    "error reading and writing download data: error",
				errUnwrap: fmt.Errorf("error"),
			},
			mocks: func() {
				mockRuntime.EXPECT().NewHTTPRequest(http.MethodGet, "", nil).Return(nil, nil)
				mockRuntime.EXPECT().DoRequest(nil).Return(&http.Response{
					StatusCode: 200,
					Body:       &fixtures.NopReadCloser{},
				}, nil)
				mockRuntime.EXPECT().CreateFile("").Return(&fixtures.NopWriteCloser{}, nil)
				mockRuntime.EXPECT().Copy(&fixtures.NopWriteCloser{}, &fixtures.NopReadCloser{}).Return(int64(0), fmt.Errorf("error"))
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// mocks
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRuntime = mockServices.NewMockFileServiceRuntime(ctrl)
			if test.mocks != nil {
				test.mocks()
			}

			fileSvc := services.NewFileService(mockRuntime)
			bytes, err := fileSvc.DownloadFile(test.input.relativePath, test.input.url, test.input.headers)
			errStr := ""
			if err != nil {
				errStr = err.Error()
			}
			assert.Equal(t, test.output, testOutput{
				bytes: bytes,
				err:   err,
			})
			assert.Equal(t, test.outProps, testOutputProps{
				errStr:    errStr,
				errUnwrap: errors.Unwrap(err),
			})
		})
	}
}

func TestRequestUploadInfoError(t *testing.T) {
	var mockRuntime *mockServices.MockFileServiceRuntime

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
					Body:       &fixtures.NopReadCloser{},
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
					Body:       &fixtures.NopReadCloser{},
				}, nil)
				mockRuntime.EXPECT().DecodeUploadInfo(&fixtures.NopReadCloser{}).Return(nil, fmt.Errorf("error"))
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// mocks
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRuntime = mockServices.NewMockFileServiceRuntime(ctrl)
			if test.mocks != nil {
				test.mocks()
			}

			fileSvc := services.NewFileService(mockRuntime)
			uploadInfo, err := fileSvc.RequestUploadInfo(test.input.authorizedServerURL, test.input.fileSize)
			errStr := ""
			if err != nil {
				errStr = err.Error()
			}
			assert.Equal(t, test.output, testOutput{
				uploadInfo: uploadInfo,
				err:        err,
			})
			assert.Equal(t, test.outProps, testOutputProps{
				errStr:    errStr,
				errUnwrap: errors.Unwrap(err),
			})
		})
	}
}

func TestUploadFileError(t *testing.T) {
	var mockRuntime *mockServices.MockFileServiceRuntime

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
				mockRuntime.EXPECT().OpenFile("").Return(&fixtures.NopReadCloser{}, nil)
				mockRuntime.EXPECT().NewHTTPRequest("", "", &fixtures.NopReadCloser{}).Return(nil, fmt.Errorf("error"))
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
				mockRuntime.EXPECT().OpenFile("").Return(&fixtures.NopReadCloser{}, nil)
				mockRuntime.EXPECT().NewHTTPRequest("", "", &fixtures.NopReadCloser{}).Return(&http.Request{}, nil)
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
				mockRuntime.EXPECT().OpenFile("").Return(&fixtures.NopReadCloser{}, nil)
				mockRuntime.EXPECT().NewHTTPRequest("", "", &fixtures.NopReadCloser{}).Return(&http.Request{}, nil)
				mockRuntime.EXPECT().DoRequest(&http.Request{}).Return(&http.Response{
					StatusCode: 201,
					Status:     "status",
					Body:       &fixtures.NopReadCloser{},
				}, nil)
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// mocks
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRuntime = mockServices.NewMockFileServiceRuntime(ctrl)
			if test.mocks != nil {
				test.mocks()
			}

			fileSvc := services.NewFileService(mockRuntime)
			err := fileSvc.UploadFile(test.input.relativePath, test.input.uploadInfo)
			errStr := ""
			if err != nil {
				errStr = err.Error()
			}
			assert.Equal(t, test.output, testOutput{
				err: err,
			})
			assert.Equal(t, test.outProps, testOutputProps{
				errStr:    errStr,
				errUnwrap: errors.Unwrap(err),
			})
		})
	}
}

func TestCompressError(t *testing.T) {
	var mockRuntime *mockServices.MockFileServiceRuntime

	type (
		testInput struct {
			inputRelativePath  string
			outputRelativePath string
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
		"create-file-for-compression-error": {
			output: testOutput{
				err: &services.CreateFileForCompressionError{Wrapped: fmt.Errorf("error")},
			},
			outProps: testOutputProps{
				errStr:    "error creating output file for compression: error",
				errUnwrap: fmt.Errorf("error"),
			},
			mocks: func() {
				mockRuntime.EXPECT().CreateFile("").Return(nil, fmt.Errorf("error"))
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// mocks
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRuntime = mockServices.NewMockFileServiceRuntime(ctrl)
			if test.mocks != nil {
				test.mocks()
			}

			fileSvc := services.NewFileService(mockRuntime)
			err := fileSvc.Compress(test.input.inputRelativePath, test.input.outputRelativePath)
			errStr := ""
			if err != nil {
				errStr = err.Error()
			}
			assert.Equal(t, test.output, testOutput{
				err: err,
			})
			assert.Equal(t, test.outProps, testOutputProps{
				errStr:    errStr,
				errUnwrap: errors.Unwrap(err),
			})
		})
	}
}

func TestVisitNodeForCompression(t *testing.T) {
	var mockRuntime *mockServices.MockFileServiceRuntime

	type (
		testInput struct {
			outTar            *tar.Writer
			inputRelativePath string
			curPath           string
			info              os.FileInfo
			err               error
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
		"walk-tree-for-compression-error": {
			input: testInput{
				err: fmt.Errorf("error"),
			},
			output: testOutput{
				err: &services.WalkTreeForCompressionError{Wrapped: fmt.Errorf("error")},
			},
			outProps: testOutputProps{
				errStr:    "error walking file tree for compression: error",
				errUnwrap: fmt.Errorf("error"),
			},
		},
		"create-compression-header-error": {
			input: testInput{
				info: &mocks.MockFileInfo{},
			},
			output: testOutput{
				err: &services.CreateCompressionHeaderError{Wrapped: fmt.Errorf("error")},
			},
			outProps: testOutputProps{
				errStr:    "error creating compression header: error",
				errUnwrap: fmt.Errorf("error"),
			},
			mocks: func() {
				curPath := filepath.Clean("")
				mockRuntime.EXPECT().CreateCompressionHeader(&mocks.MockFileInfo{}, curPath).Return(nil, fmt.Errorf("error"))
			},
		},
		"write-compression-header-error": {
			input: testInput{
				info: &mocks.MockFileInfo{},
			},
			output: testOutput{
				err: &services.WriteCompressionHeaderError{Wrapped: fmt.Errorf("error")},
			},
			outProps: testOutputProps{
				errStr:    "error writing compression header: error",
				errUnwrap: fmt.Errorf("error"),
			},
			mocks: func() {
				curPath := filepath.Clean("")
				name := filepath.ToSlash(curPath)
				mockRuntime.EXPECT().CreateCompressionHeader(&mocks.MockFileInfo{}, curPath).Return(&tar.Header{Name: name}, nil)
				mockRuntime.EXPECT().WriteCompressionHeader(nil, &tar.Header{Name: name}).Return(fmt.Errorf("error"))
			},
		},
		"open-input-file-for-compression-error": {
			input: testInput{
				info: &mocks.MockFileInfo{IsDiri: false},
			},
			output: testOutput{
				err: &services.OpenInputFileForCompressionError{Wrapped: fmt.Errorf("error")},
			},
			outProps: testOutputProps{
				errStr:    "error opening input file for compression: error",
				errUnwrap: fmt.Errorf("error"),
			},
			mocks: func() {
				curPath := filepath.Clean("")
				name := filepath.ToSlash(curPath)
				mockRuntime.EXPECT().CreateCompressionHeader(&mocks.MockFileInfo{IsDiri: false}, curPath).Return(&tar.Header{Name: name}, nil)
				mockRuntime.EXPECT().WriteCompressionHeader(nil, &tar.Header{Name: name}).Return(nil)
				mockRuntime.EXPECT().OpenFile(curPath).Return(nil, fmt.Errorf("error"))
			},
		},
		"write-input-file-for-compression-error": {
			input: testInput{
				info: &mocks.MockFileInfo{IsDiri: false},
			},
			output: testOutput{
				err: &services.WriteInputFileForCompressionError{Wrapped: fmt.Errorf("error")},
			},
			outProps: testOutputProps{
				errStr:    "error writing input file for compression: error",
				errUnwrap: fmt.Errorf("error"),
			},
			mocks: func() {
				curPath := filepath.Clean("")
				name := filepath.ToSlash(curPath)
				mockRuntime.EXPECT().CreateCompressionHeader(&mocks.MockFileInfo{IsDiri: false}, curPath).Return(&tar.Header{Name: name}, nil)
				mockRuntime.EXPECT().WriteCompressionHeader(nil, &tar.Header{Name: name}).Return(nil)
				mockRuntime.EXPECT().OpenFile(curPath).Return(&fixtures.NopReadCloser{}, nil)
				mockRuntime.EXPECT().Copy(nil, &fixtures.NopReadCloser{}).Return(int64(0), fmt.Errorf("error"))
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// mocks
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRuntime = mockServices.NewMockFileServiceRuntime(ctrl)
			if test.mocks != nil {
				test.mocks()
			}

			fileSvc := services.NewFileService(mockRuntime).(interface { // comment to skip mockgen
				VisitNodeForCompression(
					outTar *tar.Writer,
					inputRelativePath string,
					curPath string,
					info os.FileInfo,
					err error,
				) error
			})
			err := fileSvc.VisitNodeForCompression(
				test.input.outTar,
				test.input.inputRelativePath,
				test.input.curPath,
				test.input.info,
				test.input.err,
			)
			errStr := ""
			if err != nil {
				errStr = err.Error()
			}
			assert.Equal(t, test.output, testOutput{
				err: err,
			})
			assert.Equal(t, test.outProps, testOutputProps{
				errStr:    errStr,
				errUnwrap: errors.Unwrap(err),
			})
		})
	}
}
