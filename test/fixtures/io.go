package fixtures

type (
	// NopCloser implements io.Closer with no-op.
	NopCloser struct {
	}

	// NopReadCloser implements io.ReadCloser with no-op.
	NopReadCloser struct {
		NopCloser
	}

	// NopReadCloser implements io.WriteCloser with no-op.
	NopWriteCloser struct {
		NopCloser
	}
)

// Close does nothing.
func (n *NopCloser) Close() error {
	return nil
}

// Read does nothing.
func (n *NopReadCloser) Read(p []byte) (int, error) {
	return 0, nil
}

// Write does nothing.
func (n *NopWriteCloser) Write(p []byte) (int, error) {
	return 0, nil
}
