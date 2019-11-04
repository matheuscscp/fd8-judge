package uploading

import (
	"fmt"

	"github.com/matheuscscp/fd8-judge/errors"
)

type (
	// RequestUploadInfoError is a mnemonic error type that wraps an unknown error.
	RequestUploadInfoError struct {
		errors.WrapperError
	}

	// UnexpectedStatusInUploadInfoResponseError is a mnemonic error type.
	UnexpectedStatusInUploadInfoResponseError struct {
		// Status is the status string of the errored response.
		Status string
	}

	// DecodeUploadInfoError is a mnemonic error type that wraps an unknown error.
	DecodeUploadInfoError struct {
		errors.WrapperError
	}

	// OpenUploadFileError is a mnemonic error type that wraps an unknown error.
	OpenUploadFileError struct {
		errors.WrapperError
	}

	// BuildUploadRequestError is a mnemonic error type that wraps an unknown error.
	BuildUploadRequestError struct {
		errors.WrapperError
	}

	// DoUploadRequestError is a mnemonic error type that wraps an unknown error.
	DoUploadRequestError struct {
		errors.WrapperError
	}

	// UnexpectedStatusInUploadResponseError is a mnemonic error type that wraps an unknown error.
	UnexpectedStatusInUploadResponseError struct {
		// Status is the status string of the errored response.
		Status string
	}
)

// Error returns a string representation of the error.
func (e *RequestUploadInfoError) Error() string {
	return fmt.Sprintf("error requesting upload info: %s", e.Wrapped.Error())
}

// Error returns a string representation of the error.
func (e *UnexpectedStatusInUploadInfoResponseError) Error() string {
	return fmt.Sprintf("unexpected status in upload info response: %s", e.Status)
}

// Error returns a string representation of the error.
func (e *DecodeUploadInfoError) Error() string {
	return fmt.Sprintf("error decoding upload info: %s", e.Wrapped.Error())
}

// Error returns a string representation of the error.
func (e *OpenUploadFileError) Error() string {
	return fmt.Sprintf("error opening upload file: %s", e.Wrapped.Error())
}

// Error returns a string representation of the error.
func (e *BuildUploadRequestError) Error() string {
	return fmt.Sprintf("error building upload request: %s", e.Wrapped.Error())
}

// Error returns a string representation of the error.
func (e *DoUploadRequestError) Error() string {
	return fmt.Sprintf("error performing upload request: %s", e.Wrapped.Error())
}

// Error returns a string representation of the error.
func (e *UnexpectedStatusInUploadResponseError) Error() string {
	return fmt.Sprintf("unexpected status in upload response: %s", e.Status)
}
