package file

import "fmt"

type (
	// NoSuchFolderError is a mnemonic error that implments Is(target error) bool.
	NoSuchFolderError struct {
		// Path is the folder path.
		Path string
	}
)

// Error returns a string representation of the error.
func (e *NoSuchFolderError) Error() string {
	return fmt.Sprintf("no such folder: '%s'", e.Path)
}

// Is is a function to be used with errors.Is(err, target error) bool.
func (e *NoSuchFolderError) Is(target error) bool {
	_, ok := target.(*NoSuchFolderError)
	return ok
}
