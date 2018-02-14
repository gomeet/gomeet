package acl

import (
	"github.com/golang/protobuf/proto"
	"github.com/gomeet/gomeet/utils/log"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AclFunc is the pluggable function that performs Access Control List check.
//
// If error is returned, the error and `codes.PermissionDenied` will be returned to the user as well as the verbatim message.
type AclFunc func(ctx context.Context, fullMethodName string, req proto.Message) error

// ServiceAclFuncOverride allows a given gRPC service implementation to override the global `AclFunc`.
//
// If a service implements the AclFuncOverride method, it takes precedence over the `AclFunc` method,
// and will be called instead of AclFunc for all method invocations within that service.
type ServiceAclFuncOverride interface {
	AclFuncOverride(ctx context.Context, fullMethodName string, req proto.Message) error
}

// UnaryServerInterceptor returns a new unary server interceptors that validates incoming messages.
//
// Invalid messages will be rejected with `InvalidArgument` before reaching any userspace handlers.

func UnaryServerInterceptor(aclFunc AclFunc) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var err error
		if pbMsg, ok := req.(proto.Message); ok {
			if overrideSrv, ok := info.Server.(ServiceAclFuncOverride); ok {
				err = overrideSrv.AclFuncOverride(ctx, info.FullMethod, pbMsg)
			} else if aclFunc != nil {
				err = aclFunc(ctx, info.FullMethod, pbMsg)
			}
		}

		if err != nil {
			log.Error(ctx, "PermissionDenied call", err, log.Fields{"fullMethodName": info.FullMethod})
			return nil, status.Errorf(codes.PermissionDenied, err.Error())
		}

		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a new streaming server interceptors that validates incoming messages.
//
// The stage at which invalid messages will be rejected with `InvalidArgument` varies based on the
// type of the RPC. For `ServerStream` (1:m) requests, it will happen before reaching any userspace
// handlers. For `ClientStream` (n:1) or `BidiStream` (n:m) RPCs, the messages will be rejected on
// calls to `stream.Recv()`.
func StreamServerInterceptor(aclFunc AclFunc) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		var fn AclFunc
		if overrideSrv, ok := srv.(ServiceAclFuncOverride); ok {
			fn = overrideSrv.AclFuncOverride
		} else {
			fn = aclFunc
		}
		wrapper := &recvWrapper{stream, info, fn}
		return handler(srv, wrapper)
	}
}

type recvWrapper struct {
	grpc.ServerStream
	info    *grpc.StreamServerInfo
	aclFunc AclFunc
}

func (s *recvWrapper) RecvMsg(m interface{}) error {
	if err := s.ServerStream.RecvMsg(m); err != nil {
		return err
	}
	if pbMsg, ok := m.(proto.Message); ok {
		if s.aclFunc != nil {
			if err := s.aclFunc(s.Context(), s.info.FullMethod, pbMsg); err != nil {
				log.Error(s.Context(), "PermissionDenied call", err, log.Fields{"fullMethodName": s.info.FullMethod})
				return status.Errorf(codes.PermissionDenied, err.Error())
			}
		}
	}

	return nil
}
