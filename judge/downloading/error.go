package downloading

import (
	"fmt"

	"github.com/matheuscscp/fd8-judge/errors"
)

type (
	// BuildDownloadRequestError is a mnemonic error type that wraps an unknown error.
	BuildDownloadRequestError struct {
		errors.WrapperError
	}

	// DoDownloadRequestError is a mnemonic error type that wraps an unknown error.
	DoDownloadRequestError struct {
		errors.WrapperError
	}

	// UnexpectedStatusInDownloadResponseError is a mnemonic error type.
	UnexpectedStatusInDownloadResponseError struct {
		// Status is the status string of the errored response.
		Status string
	}

	// CreateFileError is a mnemonic error type that wraps an unknown error.
	CreateFileError struct {
		errors.WrapperError
	}
)

// Error returns a string representation of the error.
func (e *BuildDownloadRequestError) Error() string {
	return fmt.Sprintf("error building download request: %s", e.Wrapped.Error())
}

// Error returns a string representation of the error.
func (e *DoDownloadRequestError) Error() string {
	return fmt.Sprintf("error performing download request: %s", e.Wrapped.Error())
}

// Error returns a string representation of the error.
func (e *UnexpectedStatusInDownloadResponseError) Error() string {
	return fmt.Sprintf("unexpected status in download response: %s", e.Status)
}

// Error returns a string representation of the error.
func (e *CreateFileError) Error() string {
	return fmt.Sprintf("error creating file for downloaded data: %s", e.Wrapped.Error())
}
