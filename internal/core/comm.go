package core

import (
	pb "go-liteflow/pb"

	"google.golang.org/grpc"
)

// common grpc implement
type CoreComm struct {
	pb.UnimplementedCoreServer
}

type Service struct {
	pb.ServiceInfo
	ClientConn pb.CoreClient
}

func NewCoreClient(addr string) (cli pb.CoreClient, err error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return pb.NewCoreClient(conn), nil
}
