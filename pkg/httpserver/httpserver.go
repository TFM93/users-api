// Package httpserver implements HTTP server.
package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"time"
	log "users/pkg/logger"
)

const (
	_defaultReadTimeout     = 5 * time.Second
	_defaultWriteTimeout    = 5 * time.Second
	_defaultAddr            = ":8080"
	_defaultShutdownTimeout = 3 * time.Second
)

// Server represents a http server manager with configuration settings
type Server struct {
	server          *http.Server
	notify          chan error
	shutdownTimeout time.Duration
	l               log.Interface
}

// New handles the server initialization and connection
func New(handler http.Handler, opts ...Option) *Server {
	httpServer := &http.Server{
		Handler:      handler,
		ReadTimeout:  _defaultReadTimeout,
		WriteTimeout: _defaultWriteTimeout,
		Addr:         _defaultAddr,
	}

	s := &Server{
		server:          httpServer,
		notify:          make(chan error, 1),
		shutdownTimeout: _defaultShutdownTimeout,
	}

	for _, opt := range opts {
		opt(s)
	}
	if s.l == nil {
		s.l = log.New("")
	}
	s.start()

	return s
}

func (s *Server) start() {
	go func() {
		s.l.Info("httpServer: listening on %s", s.server.Addr)
		s.notify <- s.server.ListenAndServe()
		close(s.notify)
	}()
}

// Notify returns a channel that can be used be aware of server ListenAndServe method
func (s *Server) Notify() <-chan error {
	return s.notify
}

// Shutdown shuts down the http server
func (s *Server) Shutdown() error {
	s.l.Info("httpServer: shutting down ...")
	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	return s.server.Shutdown(ctx)
}

// Option type is used to configure the http server
type Option func(*Server)

// Port defines the listening http port
func Port(port int32) Option {
	return func(s *Server) {
		s.server.Addr = fmt.Sprintf(":%d", port)
	}
}

// ReadTimeout defines the server's read timeout
func ReadTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.server.ReadTimeout = timeout
	}
}

// WriteTimeout defines the server's write timeout
func WriteTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.server.WriteTimeout = timeout
	}
}

// ShutdownTimeout defines the server's shutdown timeout
func ShutdownTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.shutdownTimeout = timeout
	}
}

// WithLogger injects the logger dependency
func WithLogger(logger log.Interface) Option {
	return func(s *Server) {
		s.l = logger
	}
}
