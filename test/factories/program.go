package factories

import (
	"io/ioutil"
	"os"
)

// ProgramFactory creates program source codes for tests.
type ProgramFactory struct {
}

// CreateCpp11HelloWorld outputs a c++11 hello world program to the given relative path.
func (*ProgramFactory) Create(program, relativePath string) error {
	return ioutil.WriteFile(relativePath, []byte(program), os.ModePerm)
}
