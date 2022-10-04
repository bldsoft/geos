package middleware

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/go-chi/chi/v5/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var id uint64
var RequestIDHeader = "x-request-id"

func GetReqID(ctx context.Context) string {
	return middleware.GetReqID(ctx)
}

func WithReqID(ctx context.Context, reqID string) context.Context {
	return context.WithValue(ctx, middleware.RequestIDKey, reqID)
}

func RequestIDMiddleware(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
	reqID := requestID(ctx)
	return handler(WithReqID(ctx, reqID), req)
}

func requestID(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return newRequestID()
	}

	header, ok := md[RequestIDHeader]
	if !ok || len(header) == 0 {
		return newRequestID()
	}

	requestID := header[0]
	if len(requestID) == 0 {
		return newRequestID()
	}

	return requestID
}

func newRequestID() string {
	myid := atomic.AddUint64(&id, 1)
	return fmt.Sprintf("gRPC-%06d", myid)
}
