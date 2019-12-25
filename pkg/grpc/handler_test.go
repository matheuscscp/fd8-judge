// +build integration

package grpc_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	nethttp "net/http"
	"testing"
	"time"

	"github.com/matheuscscp/fd8-judge/pkg/grpc"
	"github.com/matheuscscp/fd8-judge/pkg/http"
	protos "github.com/matheuscscp/fd8-judge/test/grpc/protogen"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	googlegrpc "google.golang.org/grpc"
)

type Controller struct {
}

func (c *Controller) Register(server *googlegrpc.Server) {
	protos.RegisterTestServiceServer(server, c)
}

func (*Controller) GetGatewayRegisterFunc() grpc.GatewayRegisterFunc {
	return protos.RegisterTestServiceHandlerFromEndpoint
}

func (*Controller) SayHello(ctx context.Context, in *protos.HelloMessage) (*protos.HelloMessage, error) {
	return &protos.HelloMessage{HelloString: in.GetHelloString() + in.GetHelloString()}, nil
}

func TestGetHandlerFactory(t *testing.T) {
	emptyMiddleware := http.Middleware(
		func(
			w nethttp.ResponseWriter,
			r *nethttp.Request,
			next http.MiddlewareWrapper,
		) error {
			return next(w, r)
		},
	)

	opts := &grpc.Options{}
	opts.WithServerOptions(googlegrpc.EmptyServerOption{}).
		WithGatewayMuxOptions(runtime.WithLastMatchWins()).
		WithBeforeServer(emptyMiddleware).
		WithBeforeGateway(emptyMiddleware).
		WithBeforeRootMux(emptyMiddleware).
		WithBeforeH2C(emptyMiddleware)

	var tests = map[string]struct {
		opts *grpc.Options
	}{
		"without-options": {},
		"with-options": {
			opts: opts,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			GetHandlerFactory(t, test.opts)
		})
	}
}

func GetHandlerFactory(t *testing.T, opts *grpc.Options) {
	// configure server
	server := http.NewServer(&http.Server{
		HTTPRandomPort: true,
		HandlerFactory: grpc.GetHandlerFactory(opts, &Controller{}),
		HealthHandler: nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
			w.WriteHeader(nethttp.StatusOK)
		}),
		Logger: logrus.WithField("app", "test"),
	}, nil)

	// start
	errorChannel := make(chan error, 1)
	go func() { errorChannel <- server.Serve() }()

	err := server.WaitForReady(time.Second)
	assert.Equal(t, nil, err)

	// make gRPC request
	grpcClient, err := googlegrpc.Dial(
		server.HTTPEndpoint,
		googlegrpc.WithInsecure(),
		googlegrpc.WithBlock(),
	)
	require.Equal(t, nil, err)
	testServiceClient := protos.NewTestServiceClient(grpcClient)
	reply, err := testServiceClient.SayHello(context.Background(), &protos.HelloMessage{
		HelloString: "hello, world!",
	})
	if err != nil {
		t.Error(err)
	} else {
		assert.Equal(t, "hello, world!hello, world!", reply.GetHelloString())
	}

	// make REST request
	resp, err := nethttp.Get(
		fmt.Sprintf("http://localhost%s/hello?hello_string=hello%%2C+world%%21", server.HTTPEndpoint),
	)
	if err != nil {
		t.Error(err)
	} else {
		bytes, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			t.Error(err)
		} else {
			var result struct {
				HelloString string `json:"hello_string"`
			}
			if err := json.Unmarshal(bytes, &result); err != nil {
				t.Error(err)
			} else {
				assert.Equal(t, "hello, world!hello, world!", result.HelloString)
			}
		}
	}

	// stop
	server.GracefulShutdown()
	if err := <-errorChannel; err != nil {
		t.Error(err)
	}
}
