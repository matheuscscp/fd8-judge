// +build unit

package services_test

import (
	"archive/tar"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/matheuscscp/fd8-judge/pkg/services"
	"github.com/matheuscscp/fd8-judge/test/fixtures"
	"github.com/matheuscscp/fd8-judge/test/mocks"
	mockServices "github.com/matheuscscp/fd8-judge/test/mocks/gen/pkg/services"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestDownloadFileError(t *testing.T) {
	t.Parallel()

	var mockRuntime *mockServices.MockdefaultFileServiceRuntime

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
				mockRuntime.EXPECT().NewRequest(http.MethodGet, "", nil).Return(nil, fmt.Errorf("error"))
			},
		},
		"do-download-request-error": {
			output: testOutput{
				err: fmt.Errorf("error performing download request: %w", fmt.Errorf("error")),
			},
			mocks: func() {
				mockRuntime.EXPECT().NewRequest(http.MethodGet, "", nil).Return(nil, nil)
				mockRuntime.EXPECT().Do(nil).Return(nil, fmt.Errorf("error"))
			},
		},
		"unexpected-status-in-download-response-error": {
			output: testOutput{
				err: fmt.Errorf("unexpected status in download response: status"),
			},
			mocks: func() {
				mockRuntime.EXPECT().NewRequest(http.MethodGet, "", nil).Return(nil, nil)
				mockRuntime.EXPECT().Do(nil).Return(&http.Response{
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
				mockRuntime.EXPECT().NewRequest(http.MethodGet, "", nil).Return(nil, nil)
				mockRuntime.EXPECT().Do(nil).Return(&http.Response{
					StatusCode: 200,
					Body:       &fixtures.NopReadCloser{},
				}, nil)
				mockRuntime.EXPECT().Create("").Return(nil, fmt.Errorf("error"))
			},
		},
		"transfer-and-store-download-data-error": {
			output: testOutput{
				err: fmt.Errorf("error reading and writing download data: %w", fmt.Errorf("error")),
			},
			mocks: func() {
				mockRuntime.EXPECT().NewRequest(http.MethodGet, "", nil).Return(nil, nil)
				mockRuntime.EXPECT().Do(nil).Return(&http.Response{
					StatusCode: 200,
					Body:       &fixtures.NopReadCloser{},
				}, nil)
				mockRuntime.EXPECT().Create("").Return(&fixtures.NopWriteCloser{}, nil)
				mockRuntime.EXPECT().Copy(&fixtures.NopWriteCloser{}, &fixtures.NopReadCloser{}).Return(int64(0), fmt.Errorf("error"))
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// mocks
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRuntime = mockServices.NewMockdefaultFileServiceRuntime(ctrl)
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

func TestUploadFileError(t *testing.T) {
	t.Parallel()

	var mockRuntime *mockServices.MockdefaultFileServiceRuntime

	type (
		testInput struct {
			relativePath        string
			authorizedServerURL string
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
		"get-file-info-error": {
			output: testOutput{
				err: fmt.Errorf("error getting file infos to get size: %w", fmt.Errorf("error")),
			},
			mocks: func() {
				mockRuntime.EXPECT().Stat(filepath.Clean("")).Return(nil, fmt.Errorf("error"))
			},
		},
		"create-upload-info-request-error": {
			output: testOutput{
				err: fmt.Errorf("error creating upload info request: %w", fmt.Errorf("error")),
			},
			mocks: func() {
				mockRuntime.EXPECT().Stat(filepath.Clean("")).Return(nil, nil)
				mockRuntime.EXPECT().NewRequest(http.MethodGet, "", nil).Return(nil, fmt.Errorf("error"))
			},
		},
		"request-upload-info-error": {
			output: testOutput{
				err: fmt.Errorf("error requesting upload info: %w", fmt.Errorf("error")),
			},
			mocks: func() {
				mockRuntime.EXPECT().Stat(filepath.Clean("")).Return(&mocks.MockFileInfo{}, nil)
				mockRuntime.EXPECT().NewRequest(http.MethodGet, "", nil).Return(&http.Request{
					Header: http.Header{},
				}, nil)
				mockRuntime.EXPECT().Do(&http.Request{
					Header: http.Header{
						services.FileUploadNameHeader: []string{"."},
						services.FileUploadSizeHeader: []string{"0"},
					},
				}).Return(nil, fmt.Errorf("error"))
			},
		},
		"unexpected-status-in-upload-info-response-error": {
			output: testOutput{
				err: fmt.Errorf("unexpected status in upload info response: status"),
			},
			mocks: func() {
				mockRuntime.EXPECT().Stat(filepath.Clean("")).Return(&mocks.MockFileInfo{}, nil)
				mockRuntime.EXPECT().NewRequest(http.MethodGet, "", nil).Return(&http.Request{
					Header: http.Header{},
				}, nil)
				mockRuntime.EXPECT().Do(&http.Request{
					Header: http.Header{
						services.FileUploadNameHeader: []string{"."},
						services.FileUploadSizeHeader: []string{"0"},
					},
				}).Return(&http.Response{
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
				mockRuntime.EXPECT().Stat(filepath.Clean("")).Return(&mocks.MockFileInfo{}, nil)
				mockRuntime.EXPECT().NewRequest(http.MethodGet, "", nil).Return(&http.Request{
					Header: http.Header{},
				}, nil)
				mockRuntime.EXPECT().Do(&http.Request{
					Header: http.Header{
						services.FileUploadNameHeader: []string{"."},
						services.FileUploadSizeHeader: []string{"0"},
					},
				}).Return(&http.Response{
					StatusCode: 200,
					Status:     "status",
					Body:       &fixtures.NopReadCloser{},
				}, nil)
				var uploadInfo struct {
					Method, URL string
					Headers     http.Header
				}
				mockRuntime.EXPECT().NewDecoderDecode(&fixtures.NopReadCloser{}, &uploadInfo).Return(fmt.Errorf("error"))
			},
		},
		"open-upload-file-error": {
			output: testOutput{
				err: fmt.Errorf("error opening upload file: %w", fmt.Errorf("error")),
			},
			mocks: func() {
				mockRuntime.EXPECT().Stat(filepath.Clean("")).Return(&mocks.MockFileInfo{}, nil)
				mockRuntime.EXPECT().NewRequest(http.MethodGet, "", nil).Return(&http.Request{
					Header: http.Header{},
				}, nil)
				mockRuntime.EXPECT().Do(&http.Request{
					Header: http.Header{
						services.FileUploadNameHeader: []string{"."},
						services.FileUploadSizeHeader: []string{"0"},
					},
				}).Return(&http.Response{
					StatusCode: 200,
					Status:     "status",
					Body:       &fixtures.NopReadCloser{},
				}, nil)
				var uploadInfo struct {
					Method, URL string
					Headers     http.Header
				}
				mockRuntime.EXPECT().NewDecoderDecode(&fixtures.NopReadCloser{}, &uploadInfo).Return(nil)
				mockRuntime.EXPECT().Open(filepath.Clean("")).Return(nil, fmt.Errorf("error"))
			},
		},
		"build-upload-request-error": {
			output: testOutput{
				err: fmt.Errorf("error building upload request: %w", fmt.Errorf("error")),
			},
			mocks: func() {
				mockRuntime.EXPECT().Stat(filepath.Clean("")).Return(&mocks.MockFileInfo{}, nil)
				mockRuntime.EXPECT().NewRequest(http.MethodGet, "", nil).Return(&http.Request{
					Header: http.Header{},
				}, nil)
				mockRuntime.EXPECT().Do(&http.Request{
					Header: http.Header{
						services.FileUploadNameHeader: []string{"."},
						services.FileUploadSizeHeader: []string{"0"},
					},
				}).Return(&http.Response{
					StatusCode: 200,
					Status:     "status",
					Body:       &fixtures.NopReadCloser{},
				}, nil)
				var uploadInfo struct {
					Method, URL string
					Headers     http.Header
				}
				mockRuntime.EXPECT().NewDecoderDecode(&fixtures.NopReadCloser{}, &uploadInfo).Return(nil)
				mockRuntime.EXPECT().Open(filepath.Clean("")).Return(&fixtures.NopReadCloser{}, nil)
				mockRuntime.EXPECT().NewRequest("", "", &fixtures.NopReadCloser{}).Return(nil, fmt.Errorf("error"))
			},
		},
		"do-upload-request-error": {
			output: testOutput{
				err: fmt.Errorf("error performing upload request: %w", fmt.Errorf("error")),
			},
			mocks: func() {
				mockRuntime.EXPECT().Stat(filepath.Clean("")).Return(&mocks.MockFileInfo{}, nil)
				mockRuntime.EXPECT().NewRequest(http.MethodGet, "", nil).Return(&http.Request{
					Header: http.Header{},
				}, nil)
				mockRuntime.EXPECT().Do(&http.Request{
					Header: http.Header{
						services.FileUploadNameHeader: []string{"."},
						services.FileUploadSizeHeader: []string{"0"},
					},
				}).Return(&http.Response{
					StatusCode: 200,
					Status:     "status",
					Body:       &fixtures.NopReadCloser{},
				}, nil)
				var uploadInfo struct {
					Method, URL string
					Headers     http.Header
				}
				mockRuntime.EXPECT().NewDecoderDecode(&fixtures.NopReadCloser{}, &uploadInfo).Return(nil)
				mockRuntime.EXPECT().Open(filepath.Clean("")).Return(&fixtures.NopReadCloser{}, nil)
				mockRuntime.EXPECT().NewRequest("", "", &fixtures.NopReadCloser{}).Return(nil, nil)
				mockRuntime.EXPECT().Do(nil).Return(nil, fmt.Errorf("error"))
			},
		},
		"unexpected-status-in-upload-response-error": {
			output: testOutput{
				err: fmt.Errorf("unexpected status in upload response: status"),
			},
			mocks: func() {
				mockRuntime.EXPECT().Stat(filepath.Clean("")).Return(&mocks.MockFileInfo{}, nil)
				mockRuntime.EXPECT().NewRequest(http.MethodGet, "", nil).Return(&http.Request{
					Header: http.Header{},
				}, nil)
				mockRuntime.EXPECT().Do(&http.Request{
					Header: http.Header{
						services.FileUploadNameHeader: []string{"."},
						services.FileUploadSizeHeader: []string{"0"},
					},
				}).Return(&http.Response{
					StatusCode: 200,
					Status:     "status",
					Body:       &fixtures.NopReadCloser{},
				}, nil)
				var uploadInfo struct {
					Method, URL string
					Headers     http.Header
				}
				mockRuntime.EXPECT().NewDecoderDecode(&fixtures.NopReadCloser{}, &uploadInfo).Return(nil)
				mockRuntime.EXPECT().Open(filepath.Clean("")).Return(&fixtures.NopReadCloser{}, nil)
				mockRuntime.EXPECT().NewRequest("", "", &fixtures.NopReadCloser{}).Return(nil, nil)
				mockRuntime.EXPECT().Do(nil).Return(&http.Response{
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
			mockRuntime = mockServices.NewMockdefaultFileServiceRuntime(ctrl)
			if test.mocks != nil {
				test.mocks()
			}

			fileSvc := services.NewFileService(mockRuntime)
			err := fileSvc.UploadFile(test.input.relativePath, test.input.authorizedServerURL)
			assert.Equal(t, test.output, testOutput{
				err: err,
			})
		})
	}
}

func TestCompressError(t *testing.T) {
	t.Parallel()

	var mockRuntime *mockServices.MockdefaultFileServiceRuntime

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
				mockRuntime.EXPECT().Create(outputRelativePath).Return(nil, fmt.Errorf("error"))
			},
		},
		"walk-file-tree-error": {
			output: testOutput{
				err: fmt.Errorf("error"),
			},
			mocks: func() {
				inputRelativePath := filepath.Clean("")
				outputRelativePath := filepath.Clean("")
				mockRuntime.EXPECT().Create(outputRelativePath).Return(&fixtures.NopWriteCloser{}, nil)
				mockRuntime.EXPECT().Walk(inputRelativePath, gomock.Any()).Return(fmt.Errorf("error"))
				mockRuntime.EXPECT().RemoveAll(outputRelativePath).Return(nil)
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// mocks
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRuntime = mockServices.NewMockdefaultFileServiceRuntime(ctrl)
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

func TestVisitNodeForCompressionError(t *testing.T) {
	t.Parallel()

	var mockRuntime *mockServices.MockdefaultFileServiceRuntime

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
		"walk-file-tree-error": {
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
				mockRuntime.EXPECT().FileInfoHeader(&mocks.MockFileInfo{}, "").Return(nil, fmt.Errorf("error"))
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
				mockRuntime.EXPECT().FileInfoHeader(&mocks.MockFileInfo{}, "").Return(&tar.Header{Name: name}, nil)
				mockRuntime.EXPECT().WriteHeader(nil, &tar.Header{Name: name}).Return(fmt.Errorf("error"))
			},
		},
		"open-input-file-for-compression-error": {
			input: testInput{
				info: &mocks.MockFileInfo{},
			},
			output: testOutput{
				err: fmt.Errorf("error opening input file for compression: %w", fmt.Errorf("error")),
			},
			mocks: func() {
				name := filepath.ToSlash("")
				mockRuntime.EXPECT().FileInfoHeader(&mocks.MockFileInfo{}, "").Return(&tar.Header{Name: name}, nil)
				mockRuntime.EXPECT().WriteHeader(nil, &tar.Header{Name: name}).Return(nil)
				mockRuntime.EXPECT().Open("").Return(nil, fmt.Errorf("error"))
			},
		},
		"write-input-file-for-compression-error": {
			input: testInput{
				info: &mocks.MockFileInfo{},
			},
			output: testOutput{
				err: fmt.Errorf("error writing input file for compression: %w", fmt.Errorf("error")),
			},
			mocks: func() {
				name := filepath.ToSlash("")
				mockRuntime.EXPECT().FileInfoHeader(&mocks.MockFileInfo{}, "").Return(&tar.Header{Name: name}, nil)
				mockRuntime.EXPECT().WriteHeader(nil, &tar.Header{Name: name}).Return(nil)
				mockRuntime.EXPECT().Open("").Return(&fixtures.NopReadCloser{}, nil)
				mockRuntime.EXPECT().Copy(nil, &fixtures.NopReadCloser{}).Return(int64(0), fmt.Errorf("error"))
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// mocks
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRuntime = mockServices.NewMockdefaultFileServiceRuntime(ctrl)
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

	var mockRuntime *mockServices.MockdefaultFileServiceRuntime

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
				mockRuntime.EXPECT().Open(inputRelativePath).Return(nil, fmt.Errorf("error"))
			},
		},
		"create-compression-reader-error": {
			output: testOutput{
				err: fmt.Errorf("error creating compression reader: %w", fmt.Errorf("error")),
			},
			mocks: func() {
				inputRelativePath := filepath.Clean("")
				mockRuntime.EXPECT().Open(inputRelativePath).Return(&fixtures.NopReadCloser{}, nil)
				mockRuntime.EXPECT().NewReader(&fixtures.NopReadCloser{}).Return(nil, fmt.Errorf("error"))
			},
		},
		"create-root-path-error": {
			input: testInput{
				outputRelativePath: "./nonDotRootPath",
			},
			output: testOutput{
				err: fmt.Errorf("error creating root path for uncompression: %w", fmt.Errorf("error")),
			},
			mocks: func() {
				inputRelativePath := filepath.Clean("")
				mockRuntime.EXPECT().Open(inputRelativePath).Return(&fixtures.NopReadCloser{}, nil)
				mockRuntime.EXPECT().NewReader(&fixtures.NopReadCloser{}).Return(&fixtures.NopReadCloser{}, nil)
				mockRuntime.EXPECT().MkdirAll("nonDotRootPath").Return(fmt.Errorf("error"))
			},
		},
		"read-compression-header-error": {
			output: testOutput{
				err: fmt.Errorf("error reading compression header: %w", fmt.Errorf("error")),
			},
			mocks: func() {
				inOutRelativePath := filepath.Clean("")
				inTar := tar.NewReader(&fixtures.NopReadCloser{})
				mockRuntime.EXPECT().Open(inOutRelativePath).Return(&fixtures.NopReadCloser{}, nil)
				mockRuntime.EXPECT().NewReader(&fixtures.NopReadCloser{}).Return(&fixtures.NopReadCloser{}, nil)
				mockRuntime.EXPECT().Next(inTar).Return(nil, fmt.Errorf("error"))
				mockRuntime.EXPECT().RemoveAll(inOutRelativePath)
			},
		},
		"invalid-compression-header-name-error": {
			output: testOutput{
				err: fmt.Errorf("invalid compression header name, want relative path, got '/'"),
			},
			mocks: func() {
				inOutRelativePath := filepath.Clean("")
				inTar := tar.NewReader(&fixtures.NopReadCloser{})
				mockRuntime.EXPECT().Open(inOutRelativePath).Return(&fixtures.NopReadCloser{}, nil)
				mockRuntime.EXPECT().NewReader(&fixtures.NopReadCloser{}).Return(&fixtures.NopReadCloser{}, nil)
				mockRuntime.EXPECT().Next(inTar).Return(&tar.Header{Name: "/"}, nil)
				mockRuntime.EXPECT().RemoveAll(inOutRelativePath)
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
				mockRuntime.EXPECT().Open(inOutRelativePath).Return(&fixtures.NopReadCloser{}, nil)
				mockRuntime.EXPECT().NewReader(&fixtures.NopReadCloser{}).Return(&fixtures.NopReadCloser{}, nil)
				mockRuntime.EXPECT().Next(inTar).Return(&tar.Header{
					Name:     "name",
					Typeflag: tar.TypeDir,
				}, nil)
				mockRuntime.EXPECT().MkdirAll(curPath).Return(fmt.Errorf("error"))
				mockRuntime.EXPECT().RemoveAll(inOutRelativePath)
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
				mockRuntime.EXPECT().Open(inOutRelativePath).Return(&fixtures.NopReadCloser{}, nil)
				mockRuntime.EXPECT().NewReader(&fixtures.NopReadCloser{}).Return(&fixtures.NopReadCloser{}, nil)
				mockRuntime.EXPECT().Next(inTar).Return(&tar.Header{
					Name:     "name",
					Typeflag: tar.TypeReg,
				}, nil)
				mockRuntime.EXPECT().Create(curPath).Return(nil, fmt.Errorf("error"))
				mockRuntime.EXPECT().RemoveAll(inOutRelativePath)
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
				mockRuntime.EXPECT().Open(inOutRelativePath).Return(&fixtures.NopReadCloser{}, nil)
				mockRuntime.EXPECT().NewReader(&fixtures.NopReadCloser{}).Return(&fixtures.NopReadCloser{}, nil)
				mockRuntime.EXPECT().Next(inTar).Return(&tar.Header{
					Name:     "name",
					Typeflag: tar.TypeReg,
				}, nil)
				mockRuntime.EXPECT().Create(curPath).Return(&fixtures.NopWriteCloser{}, nil)
				mockRuntime.EXPECT().Copy(&fixtures.NopWriteCloser{}, inTar).Return(int64(0), fmt.Errorf("error"))
				mockRuntime.EXPECT().RemoveAll(inOutRelativePath)
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// mocks
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRuntime = mockServices.NewMockdefaultFileServiceRuntime(ctrl)
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

func TestRemoveFileTreeError(t *testing.T) {
	t.Parallel()

	var mockRuntime *mockServices.MockdefaultFileServiceRuntime

	type (
		testInput struct {
			relativePath string
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
		"remove-file-tree-error": {
			output: testOutput{
				err: fmt.Errorf("error removing file tree: %w", fmt.Errorf("error")),
			},
			mocks: func() {
				relativePath := filepath.Clean("")
				mockRuntime.EXPECT().RemoveAll(relativePath).Return(fmt.Errorf("error"))
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// mocks
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRuntime = mockServices.NewMockdefaultFileServiceRuntime(ctrl)
			if test.mocks != nil {
				test.mocks()
			}

			fileSvc := services.NewFileService(mockRuntime)
			err := fileSvc.RemoveFileTree(test.input.relativePath)
			assert.Equal(t, test.output, testOutput{
				err: err,
			})
		})
	}
}

func TestOpenFileError(t *testing.T) {
	t.Parallel()

	var mockRuntime *mockServices.MockdefaultFileServiceRuntime

	type (
		testInput struct {
			relativePath string
		}
		testOutput struct {
			file io.ReadCloser
			err  error
		}
	)
	var tests = map[string]struct {
		input  testInput
		output testOutput
		mocks  func()
	}{
		"open-file-error": {
			output: testOutput{
				err: fmt.Errorf("error opening file: %w", fmt.Errorf("error")),
			},
			mocks: func() {
				relativePath := filepath.Clean("")
				mockRuntime.EXPECT().Open(relativePath).Return(nil, fmt.Errorf("error"))
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// mocks
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRuntime = mockServices.NewMockdefaultFileServiceRuntime(ctrl)
			if test.mocks != nil {
				test.mocks()
			}

			fileSvc := services.NewFileService(mockRuntime)
			file, err := fileSvc.OpenFile(test.input.relativePath)
			assert.Equal(t, test.output, testOutput{
				file: file,
				err:  err,
			})
		})
	}
}

func TestCreateFileError(t *testing.T) {
	t.Parallel()

	var mockRuntime *mockServices.MockdefaultFileServiceRuntime

	type (
		testInput struct {
			relativePath string
		}
		testOutput struct {
			file io.WriteCloser
			err  error
		}
	)
	var tests = map[string]struct {
		input  testInput
		output testOutput
		mocks  func()
	}{
		"create-folder-error": {
			input: testInput{
				relativePath: "./folder/file.txt",
			},
			output: testOutput{
				err: fmt.Errorf("error creating folder for file: %w", fmt.Errorf("error")),
			},
			mocks: func() {
				folderPath := filepath.Dir(filepath.Clean("./folder/file.txt"))
				mockRuntime.EXPECT().MkdirAll(folderPath).Return(fmt.Errorf("error"))
			},
		},
		"create-file-error": {
			output: testOutput{
				err: fmt.Errorf("error creating file: %w", fmt.Errorf("error")),
			},
			mocks: func() {
				relativePath := filepath.Clean("")
				mockRuntime.EXPECT().Create(relativePath).Return(nil, fmt.Errorf("error"))
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// mocks
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRuntime = mockServices.NewMockdefaultFileServiceRuntime(ctrl)
			if test.mocks != nil {
				test.mocks()
			}

			fileSvc := services.NewFileService(mockRuntime)
			file, err := fileSvc.CreateFile(test.input.relativePath)
			assert.Equal(t, test.output, testOutput{
				file: file,
				err:  err,
			})
		})
	}
}

func TestListFilesError(t *testing.T) {
	t.Parallel()

	var mockRuntime *mockServices.MockdefaultFileServiceRuntime

	type (
		testInput struct {
			relativePath string
		}
		testOutput struct {
			files     []string
			err       error
			errString string
		}
	)
	var tests = map[string]struct {
		input  testInput
		output testOutput
		mocks  func()
	}{
		"no-such-folder-error": {
			output: testOutput{
				err:       &services.NoSuchFolderError{Path: filepath.Clean("")},
				errString: fmt.Sprintf("no such folder: '%s'", filepath.Clean("")),
			},
			mocks: func() {
				relativePath := filepath.Clean("")
				mockRuntime.EXPECT().ReadDir(relativePath).Return(nil, fmt.Errorf("no such file or directory"))
			},
		},
		"read-folder-error": {
			output: testOutput{
				err:       fmt.Errorf("error reading folder to list files: %w", fmt.Errorf("error")),
				errString: "error reading folder to list files: error",
			},
			mocks: func() {
				relativePath := filepath.Clean("")
				mockRuntime.EXPECT().ReadDir(relativePath).Return(nil, fmt.Errorf("error"))
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// mocks
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRuntime = mockServices.NewMockdefaultFileServiceRuntime(ctrl)
			if test.mocks != nil {
				test.mocks()
			}

			fileSvc := services.NewFileService(mockRuntime)
			files, err := fileSvc.ListFiles(test.input.relativePath)
			assert.Equal(t, test.output, testOutput{
				files:     files,
				err:       err,
				errString: err.Error(),
			})
		})
	}
}
