package api

import (
	"net/http"

	pb "github.com/hm-edu/domain-api"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func (s *Server) addAcmeAccount(c echo.Context) (err error) {
	conn, err := grpc.Dial("localhost:8083", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()))
	if err != nil {
		return err
	}
	client := pb.NewDomainServiceClient(conn)
	permission, err := client.CheckPermission(c.Request().Context(), &pb.CheckPermissionRequest{Domains: []string{}})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, permission)
}
