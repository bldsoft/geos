package middleware

import (
	"context"
	"strings"

	"github.com/bldsoft/gost/server/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

var RealIPHeader = "x-real-ip"

func GetRealIP(ctx context.Context) string {
	return middleware.GetRealIP(ctx)
}

func RealIPMiddleware(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
	ip := realIP(ctx)
	return handler(middleware.WithRealIP(ctx, ip), req)
}

func realIP(ctx context.Context) string {
	if ip := getHeader(ctx, RealIPHeader); ip != "" {
		return ip
	}
	p, _ := peer.FromContext(ctx)
	return strings.Split(p.Addr.String(), ":")[0]
}
