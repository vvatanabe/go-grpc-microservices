package interceptor

import (
	"context"
	"time"

	"log"

	"github.com/vvatanabe/go-grpc-microservices/shared/md"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func XTraceID() grpc.UnaryServerInterceptor {
	return func(ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {
		traceID := md.GetTraceIDFromContext(ctx)
		ctx = md.AddTraceIDToContext(ctx, traceID)
		return handler(ctx, req)
	}
}

const loggingFmt = "TraceID:%s\tFullMethod:%s\tElapsedTime:%s\tStatusCode:%s\tError:%s\n"

func Logging() grpc.UnaryServerInterceptor {
	return func(ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		h, err := handler(ctx, req)
		var errMsg string
		if err != nil {
			errMsg = err.Error()
		}
		log.Printf(loggingFmt,
			md.GetTraceIDFromContext(ctx),
			info.FullMethod,
			time.Since(start),
			status.Code(err), errMsg)
		return h, err
	}
}

func XUserID() grpc.UnaryServerInterceptor {
	return func(ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {
		userID, err := md.SafeGetUserIDFromContext(ctx)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		ctx = md.AddUserIDToContext(ctx, userID)
		return handler(ctx, req)
	}
}
