// +build unit

package services_test

import (
	"archive/tar"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/matheuscscp/fd8-judge/pkg/services"
	"github.com/matheuscscp/fd8-judge/test/fixtures"
	"github.com/matheuscscp/fd8-judge/test/mocks"
	mockServices "github.com/matheuscscp/fd8-judge/test/mocks/pkg/services"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestDownloadFileError(t *testing.T) {
	t.Parallel()

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
	)
	var tests = map[string]struct {
		input  testInput
		output testOutput
		mocks  func()
	}{
		"build-download-request-error": {
			output: testOutput{
				err: fmt.Errorf("error building download request: %w", fmt.Errorf("error")),
			},
			mocks: func() {
				mockRuntime.EXPECT().NewHTTPRequest(http.MethodGet, "", nil).Return(nil, fmt.Errorf("error"))
			},
		},
		"do-download-request-error": {
			output: testOutput{
				err: fmt.Errorf("error performing download request: %w", fmt.Errorf("error")),
			},
			mocks: func() {
				mockRuntime.EXPECT().NewHTTPRequest(http.MethodGet, "", nil).Return(nil, nil)
				mockRuntime.EXPECT().DoRequest(nil).Return(nil, fmt.Errorf("error"))
			},
		},
		"unexpected-status-in-download-response-error": {
			output: testOutput{
				err: fmt.Errorf("unexpected status in download response: status"),
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
				err: fmt.Errorf("error creating file for download data: %w", fmt.Errorf("error")),
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
				err: fmt.Errorf("error reading and writing download data: %w", fmt.Errorf("error")),
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
			assert.Equal(t, test.output, testOutput{
				bytes: bytes,
				err:   err,
			})
		})
	}
}

func TestRequestUploadInfoError(t *testing.T) {
	t.Parallel()

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
	)
	var tests = map[string]struct {
		input  testInput
		output testOutput
		mocks  func()
	}{
		"request-upload-info-error": {
			output: testOutput{
				err: fmt.Errorf("error requesting upload info: %w", fmt.Errorf("error")),
			},
			mocks: func() {
				mockRuntime.EXPECT().DoGetRequest("?fileSize=0").Return(nil, fmt.Errorf("error"))
			},
		},
		"unexpected-status-in-download-response-error": {
			output: testOutput{
				err: fmt.Errorf("unexpected status in upload info response: status"),
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
				err: fmt.Errorf("error decoding upload info: %w", fmt.Errorf("error")),
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
			assert.Equal(t, test.output, testOutput{
				uploadInfo: uploadInfo,
				err:        err,
			})
		})
	}
}

func TestUploadFileError(t *testing.T) {
	t.Parallel()

	var mockRuntime *mockServices.MockFileServiceRuntime

	type (
		testInput struct {
			relativePath string
			uploadInfo   *services.FileUploadInfo
		}
		testOutput struct {
			err error
		}
	)
	var tests = map[string]struct {
		input  testInput
		output testOutput
		mocks  func()
	}{
		"open-upload-file-error": {
			output: testOutput{
				err: fmt.Errorf("error opening upload file: %w", fmt.Errorf("error")),
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
				err: fmt.Errorf("error building upload request: %w", fmt.Errorf("error")),
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
				err: fmt.Errorf("error performing upload request: %w", fmt.Errorf("error")),
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
				err: fmt.Errorf("unexpected status in upload response: status"),
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
			assert.Equal(t, test.output, testOutput{
				err: err,
			})
		})
	}
}

func TestCompressError(t *testing.T) {
	t.Parallel()

	var mockRuntime *mockServices.MockFileServiceRuntime

	type (
		testInput struct {
			inputRelativePath  string
			outputRelativePath string
		}
		testOutput struct {
			err error
		}
	)
	var tests = map[string]struct {
		input  testInput
		output testOutput
		mocks  func()
	}{
		"create-file-for-compression-error": {
			output: testOutput{
				err: fmt.Errorf("error creating output file for compression: %w", fmt.Errorf("error")),
			},
			mocks: func() {
				outputRelativePath := filepath.Clean("")
				mockRuntime.EXPECT().CreateFile(outputRelativePath).Return(nil, fmt.Errorf("error"))
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
			assert.Equal(t, test.output, testOutput{
				err: err,
			})
		})
	}
}

func TestVisitNodeForCompression(t *testing.T) {
	t.Parallel()

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
	)
	var tests = map[string]struct {
		input  testInput
		output testOutput
		mocks  func()
	}{
		"walk-tree-for-compression-error": {
			input: testInput{
				err: fmt.Errorf("error"),
			},
			output: testOutput{
				err: fmt.Errorf("error walking file tree for compression: %w", fmt.Errorf("error")),
			},
		},
		"create-compression-header-error": {
			input: testInput{
				info: &mocks.MockFileInfo{},
			},
			output: testOutput{
				err: fmt.Errorf("error creating compression header: %w", fmt.Errorf("error")),
			},
			mocks: func() {
				mockRuntime.EXPECT().CreateCompressionHeader(&mocks.MockFileInfo{}, "").Return(nil, fmt.Errorf("error"))
			},
		},
		"write-compression-header-error": {
			input: testInput{
				info: &mocks.MockFileInfo{},
			},
			output: testOutput{
				err: fmt.Errorf("error writing compression header: %w", fmt.Errorf("error")),
			},
			mocks: func() {
				name := filepath.ToSlash("")
				mockRuntime.EXPECT().CreateCompressionHeader(&mocks.MockFileInfo{}, "").Return(&tar.Header{Name: name}, nil)
				mockRuntime.EXPECT().WriteCompressionHeader(nil, &tar.Header{Name: name}).Return(fmt.Errorf("error"))
			},
		},
		"open-input-file-for-compression-error": {
			input: testInput{
				info: &mocks.MockFileInfo{IsDiri: false},
			},
			output: testOutput{
				err: fmt.Errorf("error opening input file for compression: %w", fmt.Errorf("error")),
			},
			mocks: func() {
				name := filepath.ToSlash("")
				mockRuntime.EXPECT().CreateCompressionHeader(&mocks.MockFileInfo{IsDiri: false}, "").Return(&tar.Header{Name: name}, nil)
				mockRuntime.EXPECT().WriteCompressionHeader(nil, &tar.Header{Name: name}).Return(nil)
				mockRuntime.EXPECT().OpenFile("").Return(nil, fmt.Errorf("error"))
			},
		},
		"write-input-file-for-compression-error": {
			input: testInput{
				info: &mocks.MockFileInfo{IsDiri: false},
			},
			output: testOutput{
				err: fmt.Errorf("error writing input file for compression: %w", fmt.Errorf("error")),
			},
			mocks: func() {
				name := filepath.ToSlash("")
				mockRuntime.EXPECT().CreateCompressionHeader(&mocks.MockFileInfo{IsDiri: false}, "").Return(&tar.Header{Name: name}, nil)
				mockRuntime.EXPECT().WriteCompressionHeader(nil, &tar.Header{Name: name}).Return(nil)
				mockRuntime.EXPECT().OpenFile("").Return(&fixtures.NopReadCloser{}, nil)
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
			assert.Equal(t, test.output, testOutput{
				err: err,
			})
		})
	}
}

func TestUncompressError(t *testing.T) {
	t.Parallel()

	var mockRuntime *mockServices.MockFileServiceRuntime

	type (
		testInput struct {
			inputRelativePath  string
			outputRelativePath string
		}
		testOutput struct {
			err error
		}
	)
	var tests = map[string]struct {
		input  testInput
		output testOutput
		mocks  func()
	}{
		"open-compressed-file-error": {
			output: testOutput{
				err: fmt.Errorf("error opening compressed file: %w", fmt.Errorf("error")),
			},
			mocks: func() {
				inputRelativePath := filepath.Clean("")
				mockRuntime.EXPECT().OpenFile(inputRelativePath).Return(nil, fmt.Errorf("error"))
			},
		},
		"create-compression-reader-error": {
			output: testOutput{
				err: fmt.Errorf("error creating compression reader: %w", fmt.Errorf("error")),
			},
			mocks: func() {
				inputRelativePath := filepath.Clean("")
				mockRuntime.EXPECT().OpenFile(inputRelativePath).Return(&fixtures.NopReadCloser{}, nil)
				mockRuntime.EXPECT().CreateCompressionReader(&fixtures.NopReadCloser{}).Return(nil, fmt.Errorf("error"))
			},
		},
		"read-compression-header-error": {
			output: testOutput{
				err: fmt.Errorf("error reading compression header: %w", fmt.Errorf("error")),
			},
			mocks: func() {
				inOutRelativePath := filepath.Clean("")
				inTar := tar.NewReader(&fixtures.NopReadCloser{})
				mockRuntime.EXPECT().OpenFile(inOutRelativePath).Return(&fixtures.NopReadCloser{}, nil)
				mockRuntime.EXPECT().CreateCompressionReader(&fixtures.NopReadCloser{}).Return(&fixtures.NopReadCloser{}, nil)
				mockRuntime.EXPECT().ReadCompressionHeader(inTar).Return(nil, fmt.Errorf("error"))
				mockRuntime.EXPECT().RemoveFileTree(inOutRelativePath)
			},
		},
		"invalid-compression-header-name-error": {
			output: testOutput{
				err: fmt.Errorf("invalid compression header name, want relative path, got '/'"),
			},
			mocks: func() {
				inOutRelativePath := filepath.Clean("")
				inTar := tar.NewReader(&fixtures.NopReadCloser{})
				mockRuntime.EXPECT().OpenFile(inOutRelativePath).Return(&fixtures.NopReadCloser{}, nil)
				mockRuntime.EXPECT().CreateCompressionReader(&fixtures.NopReadCloser{}).Return(&fixtures.NopReadCloser{}, nil)
				mockRuntime.EXPECT().ReadCompressionHeader(inTar).Return(&tar.Header{Name: "/"}, nil)
				mockRuntime.EXPECT().RemoveFileTree(inOutRelativePath)
			},
		},
		"create-folder-for-uncompression-error": {
			output: testOutput{
				err: fmt.Errorf("error creating folder for uncompression: %w", fmt.Errorf("error")),
			},
			mocks: func() {
				inOutRelativePath := filepath.Clean("")
				inTar := tar.NewReader(&fixtures.NopReadCloser{})
				curPath := filepath.Join(inOutRelativePath, "name")
				mockRuntime.EXPECT().OpenFile(inOutRelativePath).Return(&fixtures.NopReadCloser{}, nil)
				mockRuntime.EXPECT().CreateCompressionReader(&fixtures.NopReadCloser{}).Return(&fixtures.NopReadCloser{}, nil)
				mockRuntime.EXPECT().ReadCompressionHeader(inTar).Return(&tar.Header{
					Name:     "name",
					Typeflag: tar.TypeDir,
				}, nil)
				mockRuntime.EXPECT().CreateFolder(curPath).Return(fmt.Errorf("error"))
				mockRuntime.EXPECT().RemoveFileTree(inOutRelativePath)
			},
		},
		"create-file-for-uncompression-error": {
			output: testOutput{
				err: fmt.Errorf("error creating file for uncompression: %w", fmt.Errorf("error")),
			},
			mocks: func() {
				inOutRelativePath := filepath.Clean("")
				inTar := tar.NewReader(&fixtures.NopReadCloser{})
				curPath := filepath.Join(inOutRelativePath, "name")
				mockRuntime.EXPECT().OpenFile(inOutRelativePath).Return(&fixtures.NopReadCloser{}, nil)
				mockRuntime.EXPECT().CreateCompressionReader(&fixtures.NopReadCloser{}).Return(&fixtures.NopReadCloser{}, nil)
				mockRuntime.EXPECT().ReadCompressionHeader(inTar).Return(&tar.Header{
					Name:     "name",
					Typeflag: tar.TypeReg,
				}, nil)
				mockRuntime.EXPECT().CreateFile(curPath).Return(nil, fmt.Errorf("error"))
				mockRuntime.EXPECT().RemoveFileTree(inOutRelativePath)
			},
		},
		"write-output-file-for-uncompression-error": {
			output: testOutput{
				err: fmt.Errorf("error writing output file for uncompression: %w", fmt.Errorf("error")),
			},
			mocks: func() {
				inOutRelativePath := filepath.Clean("")
				inTar := tar.NewReader(&fixtures.NopReadCloser{})
				curPath := filepath.Join(inOutRelativePath, "name")
				mockRuntime.EXPECT().OpenFile(inOutRelativePath).Return(&fixtures.NopReadCloser{}, nil)
				mockRuntime.EXPECT().CreateCompressionReader(&fixtures.NopReadCloser{}).Return(&fixtures.NopReadCloser{}, nil)
				mockRuntime.EXPECT().ReadCompressionHeader(inTar).Return(&tar.Header{
					Name:     "name",
					Typeflag: tar.TypeReg,
				}, nil)
				mockRuntime.EXPECT().CreateFile(curPath).Return(&fixtures.NopWriteCloser{}, nil)
				mockRuntime.EXPECT().Copy(&fixtures.NopWriteCloser{}, inTar).Return(int64(0), fmt.Errorf("error"))
				mockRuntime.EXPECT().RemoveFileTree(inOutRelativePath)
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
			err := fileSvc.Uncompress(test.input.inputRelativePath, test.input.outputRelativePath)
			assert.Equal(t, test.output, testOutput{
				err: err,
			})
		})
	}
}
