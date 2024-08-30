package http

import (
	"context"
	"fmt"
	"net/http"
	"users/internal/app"
	"users/pkg/logger"

	gen "users/gen/proto/go"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Setup creates a new gin Engine, configures the middlewares and registers the routes
func Setup(l logger.Interface, grpcServerPort int32, healthCheck app.HealthCheckQueries) (*gin.Engine, error) {
	engine := gin.New()
	engine.Use(l.GinLoggerFn())
	engine.Use(gin.Recovery())

	engine.GET("/healthz", func(c *gin.Context) { c.Status(http.StatusOK) })
	engine.GET("/readiness", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ready"}) })
	engine.GET("/liveness", func(c *gin.Context) {
		if healthCheck.Check(c.Request.Context()) {
			c.JSON(http.StatusOK, gin.H{"status": "healthy"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "unhealthy"})
	})

	mux, err := configureGRPCGateway(grpcServerPort)
	if err != nil {
		return engine, err
	}
	engine.Any("/v1/*{grpc_gateway}", gin.WrapH(mux))
	return engine, nil
}

func configureGRPCGateway(grpcServerPort int32) (*runtime.ServeMux, error) {
	mux := runtime.NewServeMux(
		runtime.WithErrorHandler(customHTTPErrorHandler),
	)
	err := gen.RegisterUserServiceHandlerFromEndpoint(context.Background(), mux, fmt.Sprintf("127.0.0.1:%d", grpcServerPort),
		[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())})
	if err != nil {
		return nil, err
	}
	return mux, nil
}

func customHTTPErrorHandler(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, writer http.ResponseWriter, request *http.Request, err error) {
	newError := runtime.HTTPStatusError{
		HTTPStatus: http.StatusBadRequest,
		Err:        err,
	}
	runtime.DefaultHTTPErrorHandler(ctx, mux, marshaler, writer, request, &newError)
}
