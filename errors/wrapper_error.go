package errors

// WrappedError wraps another error.
type WrapperError struct {
	// Wrapped is the wrapped error.
	Wrapped error
}

// Unwrap returns the wrapped error.
func (e *WrapperError) Unwrap() error {
	return e.Wrapped
}
