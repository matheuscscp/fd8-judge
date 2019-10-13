package downloading

import "fmt"

type (
	// BuildDownloadRequestError is a mnemonic error type that wraps an unknown error.
	BuildDownloadRequestError struct {
		// Wrapped holds the wrapped error.
		Wrapped error
	}

	// DoDownloadRequestError is a mnemonic error type that wraps an unknown error.
	DoDownloadRequestError struct {
		// Wrapped holds the wrapped error.
		Wrapped error
	}

	// UnexpectedStatusInDownloadResponseError is a mnemonic error type.
	UnexpectedStatusInDownloadResponseError struct {
		// Status is the status string of the errored response.
		Status string
	}

	// CreateFileError is a mnemonic error type that wraps an unknown error.
	CreateFileError struct {
		// Wrapped holds the wrapped error.
		Wrapped error
	}
)

// Error returns a string representation of the error.
func (e *BuildDownloadRequestError) Error() string {
	return fmt.Sprintf("failed to build download request: %s", e.Wrapped.Error())
}

// Unwrap returns the unknown wrapped error.
func (e *BuildDownloadRequestError) Unwrap() error {
	return e.Wrapped
}

// Error returns a string representation of the error.
func (e *DoDownloadRequestError) Error() string {
	return fmt.Sprintf("failed to do download request: %s", e.Wrapped.Error())
}

// Unwrap returns the unknown wrapped error.
func (e *DoDownloadRequestError) Unwrap() error {
	return e.Wrapped
}

// Error returns a string representation of the error.
func (e *UnexpectedStatusInDownloadResponseError) Error() string {
	return fmt.Sprintf("received unexpected status in download response: %s", e.Status)
}

// Error returns a string representation of the error.
func (e *CreateFileError) Error() string {
	return fmt.Sprintf("failed to create file to store downloaded data: %s", e.Wrapped.Error())
}

// Unwrap returns the unknown wrapped error.
func (e *CreateFileError) Unwrap() error {
	return e.Wrapped
}
