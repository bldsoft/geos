package middleware

import (
	"context"
	"runtime/debug"

	"github.com/bldsoft/gost/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func RecoveryMiddleware(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.FromContext(ctx).Errorf("Panic %v\n%s", r, debug.Stack())
			err = status.Errorf(codes.Internal, "%v", r)
		}
	}()

	return handler(ctx, req)
}
