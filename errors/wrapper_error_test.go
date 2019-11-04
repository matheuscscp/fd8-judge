// +build unit

package errors_test

import (
	"fmt"
	"testing"

	"github.com/matheuscscp/fd8-judge/errors"
	"github.com/stretchr/testify/assert"
)

func TestUnwrap(t *testing.T) {
	err := errors.WrapperError{Wrapped: fmt.Errorf("error")}
	assert.Equal(t, fmt.Errorf("error"), err.Unwrap())
}
