package factories

import (
	"path/filepath"

	"github.com/matheuscscp/fd8-judge/pkg/services/file"
)

// ProblemBundleFactory creates problem bundles for tests.
type ProblemBundleFactory struct {
}

// CreateCpp11HelloWorldBundle creates a hello world problem bundle for tests.
func (*ProblemBundleFactory) Create(bundle FileTreeNode, relativePath string) error {
	fileSvc := file.NewService(nil)

	if err := bundle.Write("."); err != nil {
		return err
	}

	if err := fileSvc.Compress("./bundle", filepath.Clean(relativePath)); err != nil {
		return err
	}

	if err := bundle.Remove("."); err != nil {
		return err
	}

	return nil
}
