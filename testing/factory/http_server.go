package factory

import (
	"fmt"
	"net"
	"net/http"
)

// HTTPServerFactory represents a factory of HTTP servers for testing.
type HTTPServerFactory struct {
}

// NewHTTPServerFactory creates a new instance of HTTPServerFactory.
func NewHTTPServerFactory() *HTTPServerFactory {
	return &HTTPServerFactory{}
}

// NewDummy creates a new TCP net.Listener at a random port, a new http.ServerMux responding
// "PAYLOAD\n" at GET /dummy and a new http.Server with the new mux. This function also creates
// a go routine that calls server.Serve(listener).
func (f *HTTPServerFactory) NewDummy() (net.Listener, *http.Server, error) {
	// create listener at a random port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, nil, err
	}

	// create server and mux
	mux := http.NewServeMux()
	server := &http.Server{
		Handler: mux,
	}
	mux.HandleFunc("/dummy", func(w http.ResponseWriter, req *http.Request) {
		const payload = "PAYLOAD"
		const bytesToBeWritten = len(payload)
		bytesWritten, err := w.Write([]byte(payload))
		if err != nil {
			panic(fmt.Errorf("error writting dummy HTTP server response: %v", err))
		}
		if bytesWritten != bytesToBeWritten {
			panic(fmt.Errorf(
				"wrong number of bytes written by dummy HTTP server, want %d, got %d",
				bytesToBeWritten,
				bytesWritten,
			))
		}
	})

	// start server
	go func() {
		err := server.Serve(listener)
		if err != http.ErrServerClosed {
			panic(fmt.Errorf("error on dummy http.Server.Serve(): %v", err))
		}
	}()

	return listener, server, nil
}
