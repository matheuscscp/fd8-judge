package services

import (
	"fmt"
)

type (
	// BuildFileDownloadRequestError is a mnemonic error type that wraps an unknown error.
	BuildFileDownloadRequestError struct {
		Wrapped error
	}

	// DoFileDownloadRequestError is a mnemonic error type that wraps an unknown error.
	DoFileDownloadRequestError struct {
		Wrapped error
	}

	// UnexpectedStatusInFileDownloadResponseError is a mnemonic error type.
	UnexpectedStatusInFileDownloadResponseError struct {
		// Status is the status string of the errored response.
		Status string
	}

	// CreateFileForDownloadError is a mnemonic error type that wraps an unknown error.
	CreateFileForDownloadError struct {
		Wrapped error
	}

	// TransferAndStoreDownloadFileError is a mnemonic error type that wraps an unknown error.
	TransferAndStoreDownloadFileError struct {
		Wrapped error
	}

	// RequestFileUploadInfoError is a mnemonic error type that wraps an unknown error.
	RequestFileUploadInfoError struct {
		Wrapped error
	}

	// UnexpectedStatusInUploadInfoResponseError is a mnemonic error type.
	UnexpectedStatusInFileUploadInfoResponseError struct {
		// Status is the status string of the errored response.
		Status string
	}

	// DecodeUploadInfoError is a mnemonic error type that wraps an unknown error.
	DecodeFileUploadInfoError struct {
		Wrapped error
	}

	// OpenUploadFileError is a mnemonic error type that wraps an unknown error.
	OpenUploadFileError struct {
		Wrapped error
	}

	// BuildFileUploadRequestError is a mnemonic error type that wraps an unknown error.
	BuildFileUploadRequestError struct {
		Wrapped error
	}

	// DoUploadRequestError is a mnemonic error type that wraps an unknown error.
	DoFileUploadRequestError struct {
		Wrapped error
	}

	// UnexpectedStatusInFileUploadResponseError is a mnemonic error type that wraps an unknown error.
	UnexpectedStatusInFileUploadResponseError struct {
		// Status is the status string of the errored response.
		Status string
	}

	// CreateFileForCompressionError is a mnemonic error type that wraps an unknown error.
	CreateFileForCompressionError struct {
		Wrapped error
	}

	// WalkTreeForCompressionError is a mnemonic error type that wraps an unknown error.
	WalkTreeForCompressionError struct {
		Wrapped error
	}

	// CreateCompressionHeaderError only occurs if another error occurred before, during the
	// file tree walk.
	CreateCompressionHeaderError struct {
		Wrapped error
	}

	// WriteCompressionHeaderError is a mnemonic error type that wraps an unknown error.
	WriteCompressionHeaderError struct {
		Wrapped error
	}

	// OpenInputFileForCompressionError is a mnemonic error type that wraps an unknown error.
	OpenInputFileForCompressionError struct {
		Wrapped error
	}

	// WriteInputFileForCompressionError is a mnemonic error type that wraps an unknown error.
	WriteInputFileForCompressionError struct {
		Wrapped error
	}
)

// Error returns a string representation of the error.
func (e *BuildFileDownloadRequestError) Error() string {
	return fmt.Sprintf("error building download request: %s", e.Wrapped.Error())
}

// Unwrap returns the wrapped error.
func (e *BuildFileDownloadRequestError) Unwrap() error {
	return e.Wrapped
}

// Error returns a string representation of the error.
func (e *DoFileDownloadRequestError) Error() string {
	return fmt.Sprintf("error performing download request: %s", e.Wrapped.Error())
}

// Unwrap returns the wrapped error.
func (e *DoFileDownloadRequestError) Unwrap() error {
	return e.Wrapped
}

// Error returns a string representation of the error.
func (e *UnexpectedStatusInFileDownloadResponseError) Error() string {
	return fmt.Sprintf("unexpected status in download response: %s", e.Status)
}

// Error returns a string representation of the error.
func (e *CreateFileForDownloadError) Error() string {
	return fmt.Sprintf("error creating file for download data: %s", e.Wrapped.Error())
}

// Unwrap returns the wrapped error.
func (e *CreateFileForDownloadError) Unwrap() error {
	return e.Wrapped
}

// Error returns a string representation of the error.
func (e *TransferAndStoreDownloadFileError) Error() string {
	return fmt.Sprintf("error reading and writing download data: %s", e.Wrapped.Error())
}

// Unwrap returns the wrapped error.
func (e *TransferAndStoreDownloadFileError) Unwrap() error {
	return e.Wrapped
}

// Error returns a string representation of the error.
func (e *RequestFileUploadInfoError) Error() string {
	return fmt.Sprintf("error requesting upload info: %s", e.Wrapped.Error())
}

// Unwrap returns the wrapped error.
func (e *RequestFileUploadInfoError) Unwrap() error {
	return e.Wrapped
}

// Error returns a string representation of the error.
func (e *UnexpectedStatusInFileUploadInfoResponseError) Error() string {
	return fmt.Sprintf("unexpected status in upload info response: %s", e.Status)
}

// Error returns a string representation of the error.
func (e *DecodeFileUploadInfoError) Error() string {
	return fmt.Sprintf("error decoding upload info: %s", e.Wrapped.Error())
}

// Unwrap returns the wrapped error.
func (e *DecodeFileUploadInfoError) Unwrap() error {
	return e.Wrapped
}

// Error returns a string representation of the error.
func (e *OpenUploadFileError) Error() string {
	return fmt.Sprintf("error opening upload file: %s", e.Wrapped.Error())
}

// Unwrap returns the wrapped error.
func (e *OpenUploadFileError) Unwrap() error {
	return e.Wrapped
}

// Error returns a string representation of the error.
func (e *BuildFileUploadRequestError) Error() string {
	return fmt.Sprintf("error building upload request: %s", e.Wrapped.Error())
}

// Unwrap returns the wrapped error.
func (e *BuildFileUploadRequestError) Unwrap() error {
	return e.Wrapped
}

// Error returns a string representation of the error.
func (e *DoFileUploadRequestError) Error() string {
	return fmt.Sprintf("error performing upload request: %s", e.Wrapped.Error())
}

// Unwrap returns the wrapped error.
func (e *DoFileUploadRequestError) Unwrap() error {
	return e.Wrapped
}

// Error returns a string representation of the error.
func (e *UnexpectedStatusInFileUploadResponseError) Error() string {
	return fmt.Sprintf("unexpected status in upload response: %s", e.Status)
}

// Error returns a string representation of the error.
func (e *CreateFileForCompressionError) Error() string {
	return fmt.Sprintf("error creating output file for compression: %s", e.Wrapped.Error())
}

// Unwrap returns the wrapped error.
func (e *CreateFileForCompressionError) Unwrap() error {
	return e.Wrapped
}

// Error returns a string representation of the error.
func (e *WalkTreeForCompressionError) Error() string {
	return fmt.Sprintf("error walking file tree for compression: %s", e.Wrapped.Error())
}

// Unwrap returns the wrapped error.
func (e *WalkTreeForCompressionError) Unwrap() error {
	return e.Wrapped
}

// Error returns a string representation of the error.
func (e *CreateCompressionHeaderError) Error() string {
	return fmt.Sprintf("error creating compression header: %s", e.Wrapped.Error())
}

// Unwrap returns the wrapped error.
func (e *CreateCompressionHeaderError) Unwrap() error {
	return e.Wrapped
}

// Error returns a string representation of the error.
func (e *WriteCompressionHeaderError) Error() string {
	return fmt.Sprintf("error writing compression header: %s", e.Wrapped.Error())
}

// Unwrap returns the wrapped error.
func (e *WriteCompressionHeaderError) Unwrap() error {
	return e.Wrapped
}

// Error returns a string representation of the error.
func (e *OpenInputFileForCompressionError) Error() string {
	return fmt.Sprintf("error opening input file for compression: %s", e.Wrapped.Error())
}

// Unwrap returns the wrapped error.
func (e *OpenInputFileForCompressionError) Unwrap() error {
	return e.Wrapped
}

// Error returns a string representation of the error.
func (e *WriteInputFileForCompressionError) Error() string {
	return fmt.Sprintf("error writing input file for compression: %s", e.Wrapped.Error())
}

// Unwrap returns the wrapped error.
func (e *WriteInputFileForCompressionError) Unwrap() error {
	return e.Wrapped
}
