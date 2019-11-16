package mocks

import (
	"os"
	"time"
)

// MockFileInfo implements os.FileInfo.
type MockFileInfo struct {
	// Namei holds name.
	Namei string

	// Sizei holds size.
	Sizei int64

	// Modei holds mode.
	Modei os.FileMode

	// ModTimei holds modtime.
	ModTimei time.Time

	// IsDiri holds isdir.
	IsDiri bool

	// Sysi holds sys.
	Sysi interface{}
}

// Name returns Name_.
func (m *MockFileInfo) Name() string {
	return m.Namei
}

// Size returns Size_.
func (m *MockFileInfo) Size() int64 {
	return m.Sizei
}

// Mode returns Mode_.
func (m *MockFileInfo) Mode() os.FileMode {
	return m.Modei
}

// ModTime returns ModTime_.
func (m *MockFileInfo) ModTime() time.Time {
	return m.ModTimei
}

// IsDir returns IsDir_.
func (m *MockFileInfo) IsDir() bool {
	return m.IsDiri
}

// Sys returns Sys_.
func (m *MockFileInfo) Sys() interface{} {
	return m.Sysi
}
