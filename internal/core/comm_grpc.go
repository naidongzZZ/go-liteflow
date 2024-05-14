package core

import (
	pb "go-liteflow/pb"
)

// common grpc implement
type Comm struct {
	pb.UnimplementedCoreServer
}
