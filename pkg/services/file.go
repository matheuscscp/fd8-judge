package services

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	// FileUploadNameHeader is the header sent to the authorized server when requesting the one-time
	// authorized upload request.
	FileUploadNameHeader = "X-File-Name"

	// FileUploadSizeHeader is the header sent to the authorized server when requesting the one-time
	// authorized upload request.
	FileUploadSizeHeader = "X-File-Size"
)

type (
	// FileService provides methods to manipulate files.
	FileService interface {
		// DownloadFile downloads a file storing it in the local file system and returns the number of
		// bytes written.
		DownloadFile(relativePath, url string, headers http.Header) (int64, error)

		// UploadFile uploads a file first requesting a one-time authorized upload request to an
		// authorized server.
		// FileUploadNameHeader and FileUploadSizeHeader are sent with the file name and size when
		// requesting the one-time authorized request.
		UploadFile(relativePath string, authorizedServerURL string) error

		// Compress compresses a file or a folder into a .tar.gz file.
		Compress(inputRelativePath, outputRelativePath string) error

		// Uncompress uncompresses a .tar.gz file to file or a folder.
		Uncompress(inputRelativePath, outputRelativePath string) error

		// RemoveFileTree removes a file tree in the given relative path.
		RemoveFileTree(relativePath string) error

		// OpenFile opens the file stored in relative path and returns an io.ReadCloser.
		OpenFile(relativePath string) (io.ReadCloser, error)

		// CreateFile creates a file in the given relative path.
		CreateFile(relativePath string) (io.WriteCloser, error)

		// ListFiles returns a list of files (folders are discarded) contained in the given path.
		ListFiles(relativePath string) ([]string, error)
	}

	defaultFileServiceRuntime interface {
		NewRequest(method, url string, body io.Reader) (*http.Request, error)
		Do(req *http.Request) (*http.Response, error)
		Create(name string) (io.WriteCloser, error)
		Copy(dst io.Writer, src io.Reader) (int64, error)
		NewDecoderDecode(r io.Reader, v interface{}) error
		Open(name string) (io.ReadCloser, error)
		RemoveAll(path string) error
		Walk(root string, walkFn filepath.WalkFunc) error
		FileInfoHeader(fi os.FileInfo, link string) (*tar.Header, error)
		WriteHeader(out *tar.Writer, hdr *tar.Header) error
		NewReader(r io.Reader) (io.ReadCloser, error)
		Next(in *tar.Reader) (*tar.Header, error)
		MkdirAll(path string) error
		ReadDir(relativePath string) ([]os.FileInfo, error)
		Stat(relativePath string) (os.FileInfo, error)
	}

	defaultFileService struct {
		runtime defaultFileServiceRuntime
	}

	fileServiceDefaultRuntime struct {
	}
)

// NewFileService returns a new instance of the default implementation of FileService.
// If nil is passed, the default FileService will be created with the default defaultFileServiceRuntime.
func NewFileService(runtime defaultFileServiceRuntime) FileService {
	if runtime == nil {
		runtime = &fileServiceDefaultRuntime{}
	}
	return &defaultFileService{runtime: runtime}
}

// DownloadFile downloads a file and stores it in the given relative path.
// The int64 return value is the number of bytes downloaded.
func (f *defaultFileService) DownloadFile(relativePath, url string, headers http.Header) (int64, error) {
	// create request object
	req, err := f.runtime.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("error building download request: %w", err)
	}
	for headerName, headerValues := range headers {
		for _, headerValue := range headerValues {
			req.Header.Add(headerName, headerValue)
		}
	}

	// do request
	resp, err := f.runtime.Do(req)
	if err != nil {
		return 0, fmt.Errorf("error performing download request: %w", err)
	}
	defer resp.Body.Close()

	// check status
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status in download response: %s", resp.Status)
	}

	// create file
	out, err := f.runtime.Create(relativePath)
	if err != nil {
		return 0, fmt.Errorf("error creating file for download data: %w", err)
	}
	defer out.Close()

	// download
	bytes, err := f.runtime.Copy(out, resp.Body)
	if err != nil {
		return 0, fmt.Errorf("error reading and writing download data: %w", err)
	}
	return bytes, nil
}

// UploadFile uploads a file first requesting a one-time authorized upload request to an
// authorized server.
// FileUploadNameHeader and FileUploadSizeHeader are sent with the file name and size when
// requesting the one-time authorized request.
func (f *defaultFileService) UploadFile(relativePath string, authorizedServerURL string) error {
	relativePath = filepath.Clean(relativePath)

	// get file info to get size
	fileInfo, err := f.runtime.Stat(relativePath)
	if err != nil {
		return fmt.Errorf("error getting file infos to get size: %w", err)
	}

	// request one-time authorized upload request
	req, err := f.runtime.NewRequest(http.MethodGet, authorizedServerURL, nil)
	if err != nil {
		return fmt.Errorf("error creating upload info request: %w", err)
	}
	req.Header.Add(FileUploadNameHeader, filepath.Base(relativePath))
	req.Header.Add(FileUploadSizeHeader, fmt.Sprintf("%v", fileInfo.Size()))
	resp, err := f.runtime.Do(req)
	if err != nil {
		return fmt.Errorf("error requesting upload info: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status in upload info response: %s", resp.Status)
	}
	var uploadInfo struct {
		Method, URL string
		Headers     http.Header
	}
	if err := f.runtime.NewDecoderDecode(resp.Body, &uploadInfo); err != nil {
		return fmt.Errorf("error decoding upload info: %w", err)
	}

	// do upload
	file, err := f.runtime.Open(relativePath)
	if err != nil {
		return fmt.Errorf("error opening upload file: %w", err)
	}
	defer file.Close()
	req, err = f.runtime.NewRequest(uploadInfo.Method, uploadInfo.URL, file)
	if err != nil {
		return fmt.Errorf("error building upload request: %w", err)
	}
	for headerName, headerValues := range uploadInfo.Headers {
		for _, headerValue := range headerValues {
			req.Header.Add(headerName, headerValue)
		}
	}
	resp, err = f.runtime.Do(req)
	if err != nil {
		return fmt.Errorf("error performing upload request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status in upload response: %s", resp.Status)
	}

	return nil
}

// Compress compresses a file or a folder into a .tar.gz file.
func (f *defaultFileService) Compress(inputRelativePath, outputRelativePath string) error {
	// open file < gzip < tar output writers
	outputRelativePath = filepath.Clean(outputRelativePath)
	outFile, err := f.runtime.Create(outputRelativePath)
	if err != nil {
		return fmt.Errorf("error creating output file for compression: %w", err)
	}
	outGzip := gzip.NewWriter(outFile)
	outTar := tar.NewWriter(outGzip)
	closeOut := func() {
		outTar.Close()
		outGzip.Close()
		outFile.Close()
	}
	removeOut := func() { // should be called right before error early returns
		closeOut()
		_ = f.runtime.RemoveAll(outputRelativePath)
	}

	// walk file tree
	inputRelativePath = filepath.Clean(inputRelativePath)
	rootName := filepath.Base(inputRelativePath)
	err = f.runtime.Walk(inputRelativePath, func(curPath string, info os.FileInfo, err error) error {
		tarRelativePath := filepath.Clean(strings.TrimPrefix(curPath, inputRelativePath))
		return f.VisitNodeForCompression(outTar, filepath.Join(rootName, tarRelativePath), curPath, info, err)
	})
	if err != nil {
		removeOut()
		return err
	}

	closeOut()
	return nil
}

// VisitNodeForCompression visits a node of a file tree writing it the compressed output.
func (f *defaultFileService) VisitNodeForCompression(
	outTar *tar.Writer,
	inputRelativePath string,
	curPath string,
	info os.FileInfo,
	err error,
) error {
	if err != nil {
		return fmt.Errorf("error walking file tree for compression: %w", err)
	}

	header, err := f.runtime.FileInfoHeader(info, curPath)
	if err != nil {
		return fmt.Errorf("error creating compression header: %w", err)
	}

	// Name within tar must have relative path information
	header.Name = inputRelativePath

	if err := f.runtime.WriteHeader(outTar, header); err != nil {
		return fmt.Errorf("error writing compression header: %w", err)
	}

	// write file
	if info.Mode().IsRegular() {
		file, err := f.runtime.Open(curPath)
		if err != nil {
			return fmt.Errorf("error opening input file for compression: %w", err)
		}
		defer file.Close()

		if _, err := f.runtime.Copy(outTar, file); err != nil {
			return fmt.Errorf("error writing input file for compression: %w", err)
		}
	}

	return nil
}

// Uncompress uncompresses a .tar.gz file to file or a folder.
func (f *defaultFileService) Uncompress(inputRelativePath, outputRelativePath string) error {
	// open file > gzip > tar input readers
	inputRelativePath = filepath.Clean(inputRelativePath)
	inFile, err := f.runtime.Open(inputRelativePath)
	if err != nil {
		return fmt.Errorf("error opening compressed file: %w", err)
	}
	inGzip, err := f.runtime.NewReader(inFile)
	if err != nil {
		inFile.Close()
		return fmt.Errorf("error creating compression reader: %w", err)
	}
	inTar := tar.NewReader(inGzip)
	closeIn := func() {
		inGzip.Close()
		inFile.Close()
	}
	outputRelativePath = filepath.Clean(outputRelativePath)
	removeOut := func() { // should be called right before error early returns
		closeIn()
		_ = f.runtime.RemoveAll(outputRelativePath)
	}

	// create root path if necessary
	if outputRelativePath != "." {
		if err := f.runtime.MkdirAll(outputRelativePath); err != nil {
			return fmt.Errorf("error creating root path for uncompression: %w", err)
		}
	}

	// walk file tree
	for {
		header, err := f.runtime.Next(inTar)
		if err == io.EOF {
			break
		} else if err != nil {
			removeOut()
			return fmt.Errorf("error reading compression header: %w", err)
		}

		// validate path
		p := header.Name
		if p == "" || strings.Contains(p, `\`) || strings.HasPrefix(p, "/") || strings.Contains(p, "../") {
			removeOut()
			return fmt.Errorf("invalid compression header name, want relative path, got '%s'", p)
		}
		curPath := filepath.Join(outputRelativePath, p)

		switch header.Typeflag {
		case tar.TypeDir: // folder
			if err := f.runtime.MkdirAll(curPath); err != nil {
				removeOut()
				return fmt.Errorf("error creating folder for uncompression: %w", err)
			}
		case tar.TypeReg: // file
			file, err := f.runtime.Create(curPath)
			if err != nil {
				removeOut()
				return fmt.Errorf("error creating file for uncompression: %w", err)
			}
			if _, err := f.runtime.Copy(file, inTar); err != nil {
				file.Close()
				removeOut()
				return fmt.Errorf("error writing output file for uncompression: %w", err)
			}
			file.Close()
		}
	}

	closeIn()
	return nil
}

// RemoveFileTree removes a file tree in the given relative path.
func (f *defaultFileService) RemoveFileTree(relativePath string) error {
	if err := f.runtime.RemoveAll(filepath.Clean(relativePath)); err != nil {
		return fmt.Errorf("error removing file tree: %w", err)
	}
	return nil
}

// OpenFile opens the file stored in relative path and returns an io.ReadCloser.
func (f *defaultFileService) OpenFile(relativePath string) (io.ReadCloser, error) {
	file, err := f.runtime.Open(filepath.Clean(relativePath))
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	return file, nil
}

// CreateFile creates a file in the given relative path.
func (f *defaultFileService) CreateFile(relativePath string) (io.WriteCloser, error) {
	relativePath = filepath.Clean(relativePath)
	folderPath := filepath.Dir(relativePath)
	if folderPath != "" && folderPath != "." {
		if err := f.runtime.MkdirAll(folderPath); err != nil {
			return nil, fmt.Errorf("error creating folder for file: %w", err)
		}
	}
	file, err := f.runtime.Create(relativePath)
	if err != nil {
		return nil, fmt.Errorf("error creating file: %w", err)
	}
	return file, nil
}

// ListFiles returns a list of files (folders are discarded) contained in the given path.
func (f *defaultFileService) ListFiles(relativePath string) ([]string, error) {
	relativePath = filepath.Clean(relativePath)
	infos, err := f.runtime.ReadDir(relativePath)
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			return nil, &NoSuchFolderError{Path: relativePath}
		}
		return nil, fmt.Errorf("error reading folder to list files: %w", err)
	}
	var files []string
	for _, info := range infos {
		if info.Mode().IsRegular() {
			files = append(files, info.Name())
		}
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i] < files[j]
	})
	return files, nil
}

func (*fileServiceDefaultRuntime) NewRequest(method, url string, body io.Reader) (*http.Request, error) {
	return http.NewRequest(method, url, body)
}

func (*fileServiceDefaultRuntime) Do(req *http.Request) (*http.Response, error) {
	return http.DefaultClient.Do(req)
}

func (*fileServiceDefaultRuntime) Create(name string) (io.WriteCloser, error) {
	return os.Create(name)
}

func (*fileServiceDefaultRuntime) Copy(dst io.Writer, src io.Reader) (int64, error) {
	return io.Copy(dst, src)
}

func (*fileServiceDefaultRuntime) NewDecoderDecode(r io.Reader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
}

func (*fileServiceDefaultRuntime) Open(name string) (io.ReadCloser, error) {
	return os.Open(name)
}

func (*fileServiceDefaultRuntime) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (*fileServiceDefaultRuntime) Walk(root string, walkFn filepath.WalkFunc) error {
	return filepath.Walk(root, walkFn)
}

func (*fileServiceDefaultRuntime) FileInfoHeader(fi os.FileInfo, link string) (*tar.Header, error) {
	return tar.FileInfoHeader(fi, link)
}

func (*fileServiceDefaultRuntime) WriteHeader(out *tar.Writer, hdr *tar.Header) error {
	return out.WriteHeader(hdr)
}

func (*fileServiceDefaultRuntime) NewReader(r io.Reader) (io.ReadCloser, error) {
	return gzip.NewReader(r)
}

func (*fileServiceDefaultRuntime) Next(in *tar.Reader) (*tar.Header, error) {
	return in.Next()
}

func (*fileServiceDefaultRuntime) MkdirAll(path string) error {
	return os.MkdirAll(path, os.ModeDir|0755)
}

func (*fileServiceDefaultRuntime) ReadDir(dirname string) ([]os.FileInfo, error) {
	return ioutil.ReadDir(dirname)
}

func (*fileServiceDefaultRuntime) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}
