package api

import (
	"github.com/gofiber/fiber/v2"
	pb "github.com/hm-edu/domain-api"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func (s *Server) addAcmeAccount(c *fiber.Ctx) (err error) {
	conn, err := grpc.Dial("localhost:8083", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()))
	if err != nil {
		return err
	}
	client := pb.NewDomainServiceClient(conn)
	permission, err := client.CheckPermission(c.UserContext(), &pb.CheckPermissionRequest{Domains: []string{}})
	if err != nil {
		return err
	}
	return c.JSON(permission)
}
