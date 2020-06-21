package http

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	nethttp "net/http"
	"regexp"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
)

type (
	// HandlerFactory is a function used to create the net/http.Handler for the main server.
	HandlerFactory func(ctx context.Context, endpoint string) (nethttp.Handler, error)

	// Server exposes HTTPEndpoint with the net/http.Handler returned by HandlerFactory,
	// HTTPSEndpoint with the same net/http.Handler and an InternalEndpoint routing GET /health to
	// HealthHandler; the response of HealthHandler is used by WaitForReady() to determine if the
	// server is ready.
	// If HTTPEndpoint is empty, the endpoint is not exposed unless HTTPRandomPort is set to true
	// (the same is valid for HTTPSEndpoint and HTTPSRandomPort).
	// If InternalEndpoint is empty, a random port p is chosen and InternalEndpoint is set to ":p".
	Server struct {
		// HTTPEndpoint is the endpoint to serve the main server.
		HTTPEndpoint string

		// HTTPRandomPort controls if a random port should be chosen to expose the HTTPEndpoint; if
		// true, a random port p is chosen and the leftmost occurrence of POSIX regex ($)|(:[0-9]*$)
		// in HTTPEndpoint is replaced by ":p".
		HTTPRandomPort bool

		// HTTPSEndpoint is the endpoint to serve the main server over TLS.
		HTTPSEndpoint string

		// HTTPSRandomPort controls if a random port should be chosen to expose the HTTPSEndpoint; if
		// true, a random port p is chosen and the leftmost occurrence of POSIX regex ($)|(:[0-9]*$)
		// in HTTPSEndpoint is replaced by ":p".
		HTTPSRandomPort bool

		// CertFile contains the TLS certificate to pass to net/http.Server.ServeTLS().
		CertFile string

		// KeyFile contains the TLS key to pass to net/http.Server.ServeTLS().
		KeyFile string

		// HandlerFactory is used to create the net/http.Handler that serves both http://HTTPEndpoint
		// and https://HTTPSEndpoint if they are set to be created.
		HandlerFactory HandlerFactory

		// InternalEndpoint is the endpoint to serve helper routes like GET /health.
		InternalEndpoint string

		// HealthHandler serves GET http://InternalEndpoint/health.
		HealthHandler nethttp.Handler

		// RegisterInternalHandlers is a hook to register custom handlers at http://InternalEndpoint,
		// except GET /health which is already required.
		RegisterInternalHandlers func(*nethttp.ServeMux)

		// Logger is used to emit logs when the starts and stops.
		Logger logrus.FieldLogger

		// runtime holds real or mocked standard library functions.
		runtime serverRuntime

		// stopChannel is the notification channel for stopping the servers.
		// Producers:
		// * at most 3 server goroutines started in Serve()
		// * GracefulShutdown()
		// Consumers:
		// * Serve()
		stopChannel chan struct{}

		// stopOnce is used to produce in stopChannel only once within GracefulShutdown().
		stopOnce sync.Once

		// settleChannel works like a barrier to prevent WaitForReady() from reading
		// s.InternalEndpoint before it has been settled.
		// Producers:
		// * Serve()
		// Consumers:
		// * WaitForReady()
		settleChannel chan struct{}

		// errorChannel is the channel to reap server errors.
		// Producers:
		// * at most 3 server goroutines started in Serve()
		// Consumers:
		// * Serve()
		errorChannel chan error

		// state represents the state of the server.
		state state

		// stateMutex guards state.
		stateMutex sync.Mutex

		httpListener     net.Listener
		httpsListener    net.Listener
		internalListener net.Listener
		server           *nethttp.Server
		internalServer   *nethttp.Server
	}

	// serverRuntime gathers the interfaces of the standard library functions used by Server.
	serverRuntime interface {
		Intn(n int) int
		Listen(network, address string) (net.Listener, error)
		Serve(serve func(net.Listener) error, l net.Listener) error
	}

	// serverDefaultRuntime implements serverRuntime calling the standard library functions.
	serverDefaultRuntime struct {
	}

	// state defines the possible states of a server.
	state int
)

const (
	idle state = iota
	starting
	serving
	stopped
)

// NewServer creates or only initializes (if server is nil) the runtime of a server with the default
// runtime if runtime is nil, and also initializes other fields that compose internal state.
func NewServer(server *Server, runtime serverRuntime) *Server {
	if runtime == nil {
		runtime = &serverDefaultRuntime{}
	}
	if server == nil {
		server = &Server{}
	}
	server.runtime = runtime
	server.stopChannel = make(chan struct{}, 4)   // there are at most 4 producers
	server.settleChannel = make(chan struct{}, 1) // there is only one producer
	server.state = idle
	return server
}

// Serve starts the server and only returns when it shuts down.
func (s *Server) Serve() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.settleEndpoints()

	if err := s.setStateStartingIfIdle(); err != nil {
		return err
	}

	if err := s.openListeners(); err != nil {
		return err
	}
	defer s.closeListeners()

	if err := s.configureServer(ctx); err != nil {
		return err
	}
	defer func() { _ = s.server.Shutdown(ctx) }()

	s.configureInternalServer()
	defer func() { _ = s.internalServer.Shutdown(ctx) }()

	s.startServers()

	s.setState(serving)

	logger := s.Logger.WithFields(logrus.Fields{
		"http-endpoint":     s.HTTPEndpoint,
		"https-endpoint":    s.HTTPSEndpoint,
		"internal-endpoint": s.InternalEndpoint,
		"http-listener":     s.httpListener,
		"https-listener":    s.httpsListener,
		"internal-listener": s.internalListener,
	})
	logger.Info("Server started.")

	// block until any of the servers or GracefulShutdown() post to stopChannel, then stop all servers
	<-s.stopChannel
	s.stopServers(ctx)

	s.setState(stopped)

	logger.Info("Server stopped.")

	return s.reapServerErrors()
}

// GracefulShutdown notifies the server to stop.
func (s *Server) GracefulShutdown() {
	s.stopOnce.Do(func() { s.stopChannel <- struct{}{} })
}

// WaitForReady waits until the server is ready to serve requests, or returns a timeout error.
func (s *Server) WaitForReady(timeout time.Duration) error {
	// return nil if state>=serving
	s.stateMutex.Lock()
	if s.state >= serving {
		s.stateMutex.Unlock()
		return nil
	}
	s.stateMutex.Unlock()

	// wait until s.InternalEndpoint can be read, or timeout
	select {
	case <-s.settleChannel:
	case <-time.After(timeout):
		return errors.New("timed out waiting for server to settle endpoints")
	}

	// we check if the server is healthy every 3ms, or timeout
	ticker := time.NewTicker(3 * time.Millisecond)
	defer ticker.Stop()
	timeoutChannel := time.After(timeout)
	for {
		select {
		case <-ticker.C:
			if s.isHealthy() {
				return nil
			}
		case <-timeoutChannel:
			return errors.New("timed out waiting for server to be healthy")
		}
	}
}

// setStateStartingIfIdle atomically sets state=serving only if state=idle.
func (s *Server) setStateStartingIfIdle() error {
	s.stateMutex.Lock()
	if s.state != idle {
		s.stateMutex.Unlock()
		return errors.New("cannot start the server more than once")
	}
	s.state = starting
	s.stateMutex.Unlock()
	return nil
}

// setState atomically sets the state.
func (s *Server) setState(state state) {
	s.stateMutex.Lock()
	s.state = state
	s.stateMutex.Unlock()
}

// settleEndpoints assigns the required random ports.
func (s *Server) settleEndpoints() {
	endpoints := []*string{}
	if s.HTTPRandomPort {
		endpoints = append(endpoints, &s.HTTPEndpoint)
	}
	if s.HTTPSRandomPort {
		endpoints = append(endpoints, &s.HTTPSEndpoint)
	}
	if s.InternalEndpoint == "" {
		endpoints = append(endpoints, &s.InternalEndpoint)
	}
	s.assignRandomPorts(endpoints)
	s.settleChannel <- struct{}{}
}

// assignRandomPorts chooses a random port p and assigns p+i to the i-th given endpoint, for i=0..n-1.
// https://en.wikipedia.org/wiki/List_of_TCP_and_UDP_port_numbers#Dynamic,_private_or_ephemeral_ports
func (s *Server) assignRandomPorts(endpoints []*string) {
	n := len(endpoints)

	// choose base port
	beg, end := 49152, 65535-(n-1)
	p := beg + s.runtime.Intn(end+1-beg)

	// replace or append port suffixes
	re := regexp.MustCompilePOSIX("(:[0-9]*$)|($)")
	for i, endpoint := range endpoints {
		*endpoint = re.ReplaceAllString(*endpoint, fmt.Sprintf("localhost:%v", p+i))
	}
}

// openListeners opens the sockets.
func (s *Server) openListeners() error {
	endpoints := []string{s.HTTPEndpoint, s.HTTPSEndpoint, s.InternalEndpoint}
	listeners := []*net.Listener{&s.httpListener, &s.httpsListener, &s.internalListener}

	re := regexp.MustCompilePOSIX(":[0-9]+$")
	for i, endpoint := range endpoints {
		if re.MatchString(endpoint) {
			var err error
			*listeners[i], err = s.runtime.Listen("tcp", endpoint)
			if err != nil {
				// close all listeners iterating backwards
				for j := i - 1; 0 <= j; j-- {
					if *listeners[j] != nil {
						(*listeners[j]).Close()
					}
				}

				return fmt.Errorf("error listening http at '%s': %w", endpoint, err)
			}
		}
	}

	return nil
}

// closeListeners closes the sockets.
func (s *Server) closeListeners() {
	listeners := []net.Listener{s.httpListener, s.httpsListener, s.internalListener}
	for _, listener := range listeners {
		if listener != nil {
			listener.Close()
		}
	}
}

// configureServer creates the net/http.Server used to serve HTTP and HTTPSEndpoints.
func (s *Server) configureServer(ctx context.Context) error {
	endpoint := s.HTTPSEndpoint
	if s.httpListener != nil {
		endpoint = s.HTTPEndpoint
	}
	handler, err := s.HandlerFactory(ctx, endpoint)
	if err != nil {
		return fmt.Errorf("error creating http server handler: %w", err)
	}

	s.server = &nethttp.Server{
		Addr:    s.HTTPEndpoint,
		Handler: handler,
	}

	return nil
}

// configureInternalServer creates the net/http.Server used to serve InternalEndpoint.
func (s *Server) configureInternalServer() {
	mux := nethttp.NewServeMux()
	mux.Handle("/health", s.HealthHandler)
	if s.RegisterInternalHandlers != nil {
		s.RegisterInternalHandlers(mux)
	}

	s.internalServer = &nethttp.Server{
		Addr:    s.InternalEndpoint,
		Handler: mux,
	}
}

// countListeners returns the number of open listeners.
func (s *Server) countListeners() int {
	listeners := []net.Listener{s.httpListener, s.httpsListener, s.internalListener}
	n := 0
	for _, listener := range listeners {
		if listener != nil {
			n++
		}
	}
	return n
}

// startServers starts the goroutines to serve each server.
func (s *Server) startServers() {
	s.errorChannel = make(chan error, s.countListeners())
	routine := func(serve func(net.Listener) error, l net.Listener) {
		s.errorChannel <- s.runtime.Serve(serve, l)
		s.stopChannel <- struct{}{}
	}
	if s.httpListener != nil {
		go routine(func(l net.Listener) error {
			return fmt.Errorf("error serving http: %w", s.server.Serve(l))
		}, s.httpListener)
	}
	if s.httpsListener != nil {
		go routine(func(l net.Listener) error {
			return fmt.Errorf("error serving https: %w", s.server.ServeTLS(l, s.CertFile, s.KeyFile))
		}, s.httpsListener)
	}
	if s.internalListener != nil {
		go routine(func(l net.Listener) error {
			return fmt.Errorf("error serving internal: %w", s.internalServer.Serve(l))
		}, s.internalListener)
	}
}

// stopServers stops the goroutines serving each server.
func (s *Server) stopServers(ctx context.Context) {
	_ = s.server.Shutdown(ctx)
	_ = s.internalServer.Shutdown(ctx)
}

// reapServerErrors reaps the errors posted by each server goroutine in errorChannel.
func (s *Server) reapServerErrors() error {
	var result error
	for i := 0; i < s.countListeners(); i++ {
		if err := <-s.errorChannel; err != nil && !errors.Is(err, nethttp.ErrServerClosed) {
			result = multierror.Append(result, err)
		}
	}
	return result
}

// isHealthy returns true if GET http://InternalEndpoint/health returns 200 OK.
func (s *Server) isHealthy() bool {
	resp, err := nethttp.Get(fmt.Sprintf("http://%s/health", s.InternalEndpoint))
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
	return err == nil && resp.StatusCode == nethttp.StatusOK
}

func (*serverDefaultRuntime) Intn(n int) int {
	return rand.Intn(n)
}

func (*serverDefaultRuntime) Listen(network, address string) (net.Listener, error) {
	return net.Listen(network, address)
}

func (*serverDefaultRuntime) Serve(serve func(net.Listener) error, l net.Listener) error {
	return serve(l)
}
