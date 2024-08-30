package grpc

import (
	"context"
	"fmt"
	gen "users/gen/proto/go"
	"users/internal/app"
	"users/pkg/logger"

	"github.com/bufbuild/protovalidate-go"
	"google.golang.org/grpc"
)

// Setup creates a grpcServer, configures the necessary interceptors and registers the following services:
// - UserServiceServer
func Setup(l logger.Interface, commands app.UserServiceCommands, queries app.UserServiceQueries) (*grpc.Server, error) {
	if l == nil || commands == nil || queries == nil {
		return nil, fmt.Errorf("invalid input parameters: logger, commands, and queries must not be nil")
	}
	server := grpc.NewServer(grpc.ChainUnaryInterceptor(loggerInterceptor(l)))
	v, err := protovalidate.New()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize validator: %w", err)
	}
	gen.RegisterUserServiceServer(server, &UserHandler{l: l, serviceCommands: commands, serviceQueries: queries, protoValidator: v})
	return server, nil
}

func loggerInterceptor(l logger.Interface) func(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)
		if err != nil {
			l.Info("gRPC method: %s, error: %v", info.FullMethod, err)
		} else {
			l.Debug("gRPC method: %s, ok", info.FullMethod)
		}
		return resp, err
	}
}
