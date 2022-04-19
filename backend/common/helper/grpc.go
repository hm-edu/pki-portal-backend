package helper

import (
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// ServerWrapper is the basic structure of the GRPC server.
type ServerWrapper interface {
	grpc.ServiceRegistrar
	reflection.ServiceInfoProvider

	Serve(net.Listener) error
	GracefulStop()
}
