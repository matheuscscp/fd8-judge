package http

import (
	"context"
	nethttp "net/http"
)

type (
	// Middleware represents an HTTP handler function that receives the next Middleware (through its
	// MiddlewareWrapper) and is allowed to control the execution flow calling this next Middleware or
	// not.
	Middleware func(nethttp.ResponseWriter, *nethttp.Request, MiddlewareWrapper) error

	// MiddlewareWrapper wraps around a Middleware.
	MiddlewareWrapper func(nethttp.ResponseWriter, *nethttp.Request) error

	// middlewareContextKey is the key type to fill context values in the http package.
	middlewareContextKey string
)

const (
	// middlewareErrorContextKey is the key used to return an error from the last handler to the
	// middlewares.
	middlewareErrorContextKey middlewareContextKey = "error"
)

// ChainMiddlewares returns a net/http.Handler that chains a list of Middlewares where the last
// middleware is a simple wrapper around the given lastHandler.
func ChainMiddlewares(lastHandler nethttp.Handler, middlewares ...Middleware) nethttp.Handler {
	chain := MiddlewareWrapper(func(w nethttp.ResponseWriter, r *nethttp.Request) error {
		var err error
		ctxWithErrorPointer := context.WithValue(r.Context(), middlewareErrorContextKey, &err)
		reqWithErrorPointer := r.WithContext(ctxWithErrorPointer)
		lastHandler.ServeHTTP(w, reqWithErrorPointer)
		return err
	})
	for i := len(middlewares) - 1; 0 <= i; i-- {
		head := middlewares[i]
		tail := chain // prevents infinite recursion
		chain = MiddlewareWrapper(func(w nethttp.ResponseWriter, r *nethttp.Request) error {
			return head(w, r, tail)
		})
	}
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		_ = chain(w, r)
	})
}

// SetErrorForMiddlewares tries to return the given error for the chained middlewares and returns
// true if it was possible.
func SetErrorForMiddlewares(ctx context.Context, err error) bool {
	v := ctx.Value(middlewareErrorContextKey)
	if v == nil {
		return false
	}

	*(v.(*error)) = err
	return true
}
