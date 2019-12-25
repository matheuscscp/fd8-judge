// +build integration

package http_test

import (
	"context"
	"crypto/tls"
	"fmt"
	nethttp "net/http"
	"testing"
	"time"

	"github.com/matheuscscp/fd8-judge/pkg/http"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func Test(t *testing.T) {
	// configure server
	server := http.NewServer(nil, nil)
	server.HTTPRandomPort = true
	server.HTTPSRandomPort = true
	server.CertFile = "../../test/tls/cert-file.crt"
	server.KeyFile = "../../test/tls/key-file.key"
	server.HandlerFactory = func(context.Context, string) (nethttp.Handler, error) {
		return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
			w.WriteHeader(nethttp.StatusOK)
		}), nil
	}
	server.HealthHandler = nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		w.WriteHeader(nethttp.StatusOK)
	})
	server.Logger = logrus.WithField("app", "test")

	// start
	errorChannel := make(chan error, 1)
	go func() { errorChannel <- server.Serve() }()

	err := server.WaitForReady(time.Second)
	assert.Equal(t, nil, err)

	// make http request
	resp, err := nethttp.Get(fmt.Sprintf("http://localhost%s", server.HTTPEndpoint))
	if err != nil {
		t.Error(err)
	} else {
		resp.Body.Close()
		assert.Equal(t, nethttp.StatusOK, resp.StatusCode)
	}

	// make https request
	tr := &nethttp.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &nethttp.Client{Transport: tr}
	resp, err = client.Get(fmt.Sprintf("https://localhost%s", server.HTTPSEndpoint))
	if err != nil {
		t.Error(err)
	} else {
		resp.Body.Close()
		assert.Equal(t, nethttp.StatusOK, resp.StatusCode)
	}

	// stop
	server.GracefulShutdown()
	if err := <-errorChannel; err != nil {
		t.Error(err)
	}
}
