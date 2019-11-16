// +build integration

package services_test

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/matheuscscp/fd8-judge/pkg/services"
	"github.com/stretchr/testify/assert"
)

func TestCompileAndExecuteCpp11(t *testing.T) {
	cppSvc, err := services.NewProgramService("c++11", nil)
	assert.Equal(t, nil, err)

	err = ioutil.WriteFile("./c++11SourceCodeTest.cpp", []byte(`
#include <iostream>

using namespace std;

int main() {
	cout << "hello, world!" << endl;
	return 0;
}
`), os.ModePerm)
	assert.Equal(t, nil, err)

	err = cppSvc.Compile(context.TODO(), "./c++11SourceCodeTest.cpp", "./c++11Program")
	assert.Equal(t, nil, err)

	err = os.Remove("./c++11SourceCodeTest.cpp")
	assert.Equal(t, nil, err)

	cmd := cppSvc.GetExecutionCommand(context.TODO(), "./c++11SourceCodeTest.cpp", "./c++11Program")
	output, err := cmd.CombinedOutput()
	assert.Equal(t, nil, err)
	assert.Equal(t, []byte("hello, world!\n"), output)

	err = os.Remove("./c++11Program")
	assert.Equal(t, nil, err)
}
