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

// Name returns Namei.
func (m *MockFileInfo) Name() string {
	return m.Namei
}

// Size returns Sizei.
func (m *MockFileInfo) Size() int64 {
	return m.Sizei
}

// Mode returns Modei.
func (m *MockFileInfo) Mode() os.FileMode {
	return m.Modei
}

// ModTime returns ModTimei.
func (m *MockFileInfo) ModTime() time.Time {
	return m.ModTimei
}

// IsDir returns IsDiri.
func (m *MockFileInfo) IsDir() bool {
	return m.IsDiri
}

// Sys returns Sysi.
func (m *MockFileInfo) Sys() interface{} {
	return m.Sysi
}
