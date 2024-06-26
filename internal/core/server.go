package core

import (
	"context"
)

type Server interface {
	Start(ctx context.Context)
	Stop(ctx context.Context)
}