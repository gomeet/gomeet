package service

import (
	"fmt"
	"strings"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	grpc_acl "github.com/gomeet/gomeet/utils/grpc-middlewares/acl"
	grpc_cid "github.com/gomeet/gomeet/utils/grpc-middlewares/cid"
)

const INPROCESS_ADDRESS = "inprocgrpc"

func FixSvcAddress(s string) string {
	s = strings.ToLower(strings.Trim(s, " "))
	if s == "" {
		return INPROCESS_ADDRESS
	}

	return s
}

func recoverFromPanic(p interface{}) error {
	log.Infof("recovering from panic: %s", p)

	return fmt.Errorf("panic in gRPC procedure: %v", p)
}

func Interceptors() (grpc.StreamServerInterceptor, grpc.UnaryServerInterceptor) {
	//middlewares definition
	recoveryOpt := grpc_recovery.WithRecoveryHandler(recoverFromPanic)
	logrusEntry := log.NewEntry(log.StandardLogger())
	// stream middlewares
	sInterceptors := grpc_middleware.ChainStreamServer(
		grpc_prometheus.StreamServerInterceptor,
		grpc_cid.StreamServerInterceptor(false),
		grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
		grpc_logrus.StreamServerInterceptor(logrusEntry),
		grpc_auth.StreamServerInterceptor(nil),
		grpc_validator.StreamServerInterceptor(),
		grpc_acl.StreamServerInterceptor(nil),
		grpc_recovery.StreamServerInterceptor(recoveryOpt),
	)
	// unary middlewares
	uInterceptors := grpc_middleware.ChainUnaryServer(
		grpc_prometheus.UnaryServerInterceptor,
		grpc_cid.UnaryServerInterceptor(false),
		grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
		grpc_logrus.UnaryServerInterceptor(logrusEntry),
		grpc_auth.UnaryServerInterceptor(nil),
		grpc_validator.UnaryServerInterceptor(),
		grpc_acl.UnaryServerInterceptor(nil),
		grpc_recovery.UnaryServerInterceptor(recoveryOpt),
	)

	return sInterceptors, uInterceptors
}
