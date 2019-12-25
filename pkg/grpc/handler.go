package grpc

import (
	"context"
	"fmt"
	nethttp "net/http"
	"strings"

	"github.com/matheuscscp/fd8-judge/pkg/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	googlegrpc "google.golang.org/grpc"
)

type (
	// GatewayRegisterFunc is a function that registers a gRPC service in a gRPC Gateway ServeMux.
	GatewayRegisterFunc func(
		ctx context.Context,
		mux *runtime.ServeMux,
		endpoint string,
		opts []googlegrpc.DialOption,
	) error

	// Registerable represents an object that knows to register itself in a gRPC server and its
	// gRPC Gateway.
	Registerable interface {
		// Register registers the Registerable in the given google/grpc.Server.
		Register(server *googlegrpc.Server)

		// GetGatewayRegisterFunc returns the GatewayRegisterFunc of the Registerable.
		GetGatewayRegisterFunc() GatewayRegisterFunc
	}

	// Options holds the options to get a pkg/http.HandlerFactory.
	Options struct {
		server                   []googlegrpc.ServerOption
		gatewayMux               []runtime.ServeMuxOption
		beforeH2CMiddlewares     []http.Middleware
		beforeRootMuxMiddlewares []http.Middleware
		beforeServerMiddlewares  []http.Middleware
		beforeGatewayMiddlewares []http.Middleware
	}
)

// GetHandlerFactory takes (possibly nil) options and a list of Registerables and returns a
// pkg/http.HandlerFactory to create a net/http.Handler that serves both a gRPC server and its
// gRPC Gateway.
func GetHandlerFactory(
	opts *Options,
	registerables ...Registerable,
) http.HandlerFactory {
	if opts == nil {
		opts = &Options{}
	}
	return func(ctx context.Context, endpoint string) (nethttp.Handler, error) {
		// build grpc server handler
		grpcServer := googlegrpc.NewServer(opts.server...)
		for _, registerable := range registerables {
			registerable.Register(grpcServer)
		}
		grpcServerWithMiddlewares := http.ChainMiddlewares(grpcServer, opts.beforeServerMiddlewares...)

		// build grpc gateway handler
		gatewayMux := runtime.NewServeMux(opts.gatewayMux...)
		gatewayEndpoint := "localhost:" + endpoint[strings.LastIndex(endpoint, ":")+1:]
		gatewayOpts := []googlegrpc.DialOption{googlegrpc.WithInsecure()}
		for _, registerable := range registerables {
			gatewayRegisterFunc := registerable.GetGatewayRegisterFunc()
			if err := gatewayRegisterFunc(ctx, gatewayMux, gatewayEndpoint, gatewayOpts); err != nil {
				return nil, fmt.Errorf("error registering to gRPC gateway: %w", err)
			}
		}
		gatewayMuxWithMiddlewares := http.ChainMiddlewares(gatewayMux, opts.beforeGatewayMiddlewares...)

		// root mux decides if the request is for the server or for the gateway
		// https://github.com/philips/grpc-gateway-example/blob/a269bcb5931ca92be0ceae6130ac27ae89582ecc/cmd/serve.go#L55
		rootMux := nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
			if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
				grpcServerWithMiddlewares.ServeHTTP(w, r)
			} else {
				gatewayMuxWithMiddlewares.ServeHTTP(w, r)
			}
		})
		rootMuxWithMiddlewares := http.ChainMiddlewares(rootMux, opts.beforeRootMuxMiddlewares...)

		// h2c is required for grpc server
		h2cHandler := h2c.NewHandler(rootMuxWithMiddlewares, &http2.Server{})
		h2cWithMiddlewares := http.ChainMiddlewares(h2cHandler, opts.beforeH2CMiddlewares...)

		return h2cWithMiddlewares, nil
	}
}

// WithServerOptions appends a list of gRPC Server options to the Options object.
func (o *Options) WithServerOptions(opts ...googlegrpc.ServerOption) *Options {
	o.server = append(o.server, opts...)
	return o
}

// WithGatewayMuxOptions appends a list of gRPC Gateway ServeMux options to the Options object.
func (o *Options) WithGatewayMuxOptions(opts ...runtime.ServeMuxOption) *Options {
	o.gatewayMux = append(o.gatewayMux, opts...)
	return o
}

// WithBeforeH2C appends a list of http.Middlewares to the Options object that will be executed
// before everything else.
func (o *Options) WithBeforeH2C(middlewares ...http.Middleware) *Options {
	o.beforeH2CMiddlewares = append(o.beforeH2CMiddlewares, middlewares...)
	return o
}

// WithBeforeRootMux appends a list of http.Middlewares to the Options object that will be executed
// after the H2C handler and before the routing to gRPC server or gRPC Gateway.
func (o *Options) WithBeforeRootMux(middlewares ...http.Middleware) *Options {
	o.beforeRootMuxMiddlewares = append(o.beforeRootMuxMiddlewares, middlewares...)
	return o
}

// WithBeforeServer appends a list of http.Middlewares to the Options object that will be executed
// right before the gRPC server handler.
func (o *Options) WithBeforeServer(middlewares ...http.Middleware) *Options {
	o.beforeServerMiddlewares = append(o.beforeServerMiddlewares, middlewares...)
	return o
}

// WithBeforeGateway appends a list of http.Middlewares to the Options object that will be
// executed right before the gRPC gateway handler.
func (o *Options) WithBeforeGateway(middlewares ...http.Middleware) *Options {
	o.beforeGatewayMiddlewares = append(o.beforeGatewayMiddlewares, middlewares...)
	return o
}
