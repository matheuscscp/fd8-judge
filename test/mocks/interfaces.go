package mocks

import "net"

type (
	// Listener is a copy of net/http.Listener to generate a mock implementation.
	Listener interface {
		Accept() (net.Conn, error)
		Close() error
		Addr() net.Addr
	}
)
