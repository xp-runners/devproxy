package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

const timeout = 5 * time.Second

type server struct {
	srv    *http.Server
	listen net.Listener
}

func newServer(handler http.Handler, port int) (*server, error) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("Could not listen: %v\n", err)
	}

	s := &http.Server{
		Handler:      handler,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
	}
	return &server{s, l}, nil
}

// Serve calls the underlying HTTP server's Serve() function, passing
// it the listener
func (s server) Serve() error {
	return s.srv.Serve(s.listen)
}

// ServeTLS calls the underlying HTTP server's ServeTLS() function, passing
// it the listener
func (s server) ServeTLS(certFile, keyFile string) error {
	return s.srv.ServeTLS(s.listen, certFile, keyFile)
}

// Shutdown calls the underlying HTTP server's Shutdown() function
// with a given timeout
func (s server) Shutdown(timeout time.Duration) error {
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	defer s.listen.Close()
	return s.srv.Shutdown(ctx)
}
