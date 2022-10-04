package middleware

import (
	"context"

	"github.com/bldsoft/gost/log"
	"github.com/go-chi/chi/v5/middleware"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
)

func LoggerMiddleware() grpc.UnaryServerInterceptor {
	return grpc_middleware.ChainUnaryServer(injectLoggerMiddleware, logRequest)
}

func injectLoggerMiddleware(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
	reqID := middleware.GetReqID(ctx)

	logFields := log.Fields{log.ReqIdFieldName: reqID}
	logger := log.Logger.WithFields(logFields)

	ctx = context.WithValue(ctx, log.LoggerCtxKey, logger)
	return handler(ctx, req)
}

func logRequest(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
	log.FromContext(ctx).InfoWithFields(log.Fields{"msg": req}, "REQUEST")
	resp, err := handler(ctx, req)
	log.FromContext(ctx).InfoOrErrorWithFields(err, log.Fields{"msg": resp}, "RESPONSE")
	return resp, err
}
