// +build unit

package http_test

import (
	"context"
	"errors"
	nethttp "net/http"
	"testing"

	"github.com/matheuscscp/fd8-judge/pkg/http"

	"github.com/stretchr/testify/assert"
)

func TestChainMiddlewares(t *testing.T) {
	outputs := []interface{}{}

	lastHandler := nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		outputs = append(outputs, "last handler")
		ok := http.SetErrorForMiddlewares(r.Context(), errors.New("error"))
		assert.Equal(t, true, ok)
	})

	firstMiddleware := http.Middleware(func(w nethttp.ResponseWriter, r *nethttp.Request, next http.MiddlewareWrapper) error {
		outputs = append(outputs, "first before")
		err := next(w, r)
		outputs = append(outputs, "first after")
		outputs = append(outputs, err)
		return err
	})

	secondMiddleware := http.Middleware(func(w nethttp.ResponseWriter, r *nethttp.Request, next http.MiddlewareWrapper) error {
		outputs = append(outputs, "second before")
		err := next(w, r)
		outputs = append(outputs, "second after")
		return err
	})

	chain := http.ChainMiddlewares(lastHandler, firstMiddleware, secondMiddleware)
	chain.ServeHTTP(nil, &nethttp.Request{})
	assert.Equal(t, []interface{}{
		"first before",
		"second before",
		"last handler",
		"second after",
		"first after",
		errors.New("error"),
	}, outputs)
}

func TestSetErrorForMiddlewares(t *testing.T) {
	ok := http.SetErrorForMiddlewares(context.Background(), nil)
	assert.Equal(t, false, ok)
}
