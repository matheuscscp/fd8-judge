package factories

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/matheuscscp/fd8-judge/pkg/services"
)

// HTTPServerFactory represents a factory of HTTP servers for testing.
type HTTPServerFactory struct {
}

// NewDummy creates a new TCP net.Listener at a random port, a new http.ServerMux responding
// "PAYLOAD" at GET /dummy and a new http.Server with the new mux. This function also creates
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
	mux.HandleFunc("/dummy", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if _, err := w.Write([]byte("PAYLOAD")); err != nil {
			panic(fmt.Errorf("error writing dummy HTTP server response: %w", err))
		}
	})

	go f.serveAndPanicOnError(server, listener, "dummy")

	return listener, server, nil
}

// NewDummyUploader creates a new TCP net.Listener at a random port, a new http.ServerMux responding
// upload routes and a new http.Server with the new mux. This function also creates
// a go routine that calls server.Serve(listener).
func (f *HTTPServerFactory) NewDummyUploader() (net.Listener, *http.Server, error) {
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
	mux.HandleFunc("/upload-info", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		port := listener.Addr().(*net.TCPAddr).Port
		uploadInfo := &struct {
			Method, URL string
			Headers     http.Header
		}{
			Method: http.MethodPut,
			URL:    fmt.Sprintf("http://localhost:%d/upload", port),
			Headers: http.Header{
				"Content-Length": []string{r.Header.Get(services.FileUploadSizeHeader)},
			},
		}
		payload, err := json.Marshal(uploadInfo)
		if err != nil {
			panic(fmt.Errorf("error marshaling dummy upload HTTP server response: %w", err))
		}
		if _, err := w.Write(payload); err != nil {
			panic(fmt.Errorf("error writing dummy upload HTTP server response: %w", err))
		}
	})
	var uploadedData []byte
	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			var err error
			uploadedData, err = ioutil.ReadAll(r.Body)
			if err != nil {
				panic(fmt.Errorf("error reading upload: %w", err))
			}
			w.WriteHeader(http.StatusOK)
		} else if r.Method == http.MethodGet {
			if _, err := w.Write(uploadedData); err != nil {
				panic(fmt.Errorf("error writing upload: %w", err))
			}
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	go f.serveAndPanicOnError(server, listener, "dummy uploader")

	return listener, server, nil
}

// NewFileServer returns a server that serves files.
func (f *HTTPServerFactory) NewFileServer(rootRelativePath string) (net.Listener, *http.Server, error) {
	// create listener at a random port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, nil, err
	}

	// create server and mux
	mux := &http.ServeMux{}
	mux.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		path := r.URL.Query()["path"][0]
		file, err := os.Open(path)
		if err != nil {
			panic(fmt.Errorf("error opening requested file: %w", err))
		}
		defer file.Close()
		w.WriteHeader(http.StatusOK)
		if _, err := io.Copy(w, file); err != nil {
			panic(fmt.Errorf("error copying requested file: %w", err))
		}
	})
	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			port := listener.Addr().(*net.TCPAddr).Port
			path := filepath.Join(rootRelativePath, r.Header.Get(services.FileUploadNameHeader))
			uploadInfo := &struct {
				Method, URL string
				Headers     http.Header
			}{
				Method: http.MethodPut,
				URL:    fmt.Sprintf("http://localhost:%d/upload?path=%s", port, path),
			}
			payload, err := json.Marshal(uploadInfo)
			if err != nil {
				panic(fmt.Errorf("error marshaling file server upload info response: %w", err))
			}
			if _, err := w.Write(payload); err != nil {
				panic(fmt.Errorf("error writing file server upload info response: %w", err))
			}
		case http.MethodPut:
			file, err := os.Create(r.URL.Query()["path"][0])
			if err != nil {
				panic(fmt.Errorf("error creating file for upload in file server: %w", err))
			}
			defer file.Close()
			if _, err := io.Copy(file, r.Body); err != nil {
				panic(fmt.Errorf("error copying uploaded file for file server: %w", err))
			}
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	server := &http.Server{
		Handler: mux,
	}

	go f.serveAndPanicOnError(server, listener, "dummy")

	return listener, server, nil
}

func (f *HTTPServerFactory) serveAndPanicOnError(
	server *http.Server,
	listener net.Listener,
	serverName string,
) {
	if err := server.Serve(listener); err != http.ErrServerClosed {
		panic(fmt.Errorf("error on %s http.Server.Serve(): %w", serverName, err))
	}
}
