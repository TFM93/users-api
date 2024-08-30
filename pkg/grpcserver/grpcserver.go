// Package grpcserver implements GRPC server.
package grpcserver

import (
	"fmt"
	"net"
	log "users/pkg/logger"

	"google.golang.org/grpc"
)

const (
	_defaultAddr = ":8081"
)

// Server represents a grpc server manager with configuration settings
type Server struct {
	addr   string
	server *grpc.Server
	notify chan error
	l      log.Interface
}

// New handles the server initialization and connection
func New(server *grpc.Server, opts ...Option) *Server {
	s := &Server{
		notify: make(chan error, 1),
		addr:   _defaultAddr,
		server: server,
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
		l, err := net.Listen("tcp", s.addr)
		if err != nil {
			s.notify <- err
		}
		s.l.Info("grpcServer: listening on %s", s.addr)
		s.notify <- s.server.Serve(l)
		close(s.notify)
	}()
}

// Notify returns a channel that can be used be aware of server Listen and Serve methods
func (s *Server) Notify() <-chan error {
	return s.notify
}

// GracefulStop shuts down the grpc server
func (s *Server) GracefulStop() {
	s.l.Info("grpcServer: stopping ...")
	s.server.GracefulStop()
}

// Option type is used to configure the grpc server
type Option func(*Server)

// Port defines the listening grpc port
func Port(port int32) Option {
	return func(s *Server) {
		s.addr = fmt.Sprintf(":%d", port)
	}
}

// WithLogger injects the logger dependency
func WithLogger(logger log.Interface) Option {
	return func(s *Server) {
		s.l = logger
	}
}
