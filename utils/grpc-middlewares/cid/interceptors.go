package cid

import (
	"github.com/google/uuid"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// UnaryServerInterceptor returns a new unary server interceptors that performs per-request.
func UnaryServerInterceptor(forceNew bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		return handler(ensureCorrelationId(ctx, forceNew), req)
	}
}

// StreamServerInterceptor returns a new unary server interceptors that performs per-request auth.
func StreamServerInterceptor(forceNew bool) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		newCtx := ensureCorrelationId(stream.Context(), forceNew)
		wrapped := grpc_middleware.WrapServerStream(stream)
		wrapped.WrappedContext = newCtx
		return handler(srv, wrapped)
	}
}

func ensureCorrelationId(ctx context.Context, forceNew bool) context.Context {
	var cid string
	nmd := metautils.ExtractIncoming(ctx)
	if !forceNew {
		cid = nmd.Get(correlationId)
	}
	if cid == "" {
		nmd.Add(correlationId, uuid.New().String())
		nCtx := nmd.ToIncoming(ctx)
		return nCtx
	}

	return ctx
}
