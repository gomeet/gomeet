package log

import (
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"

	grpc_cid "github.com/gomeet/gomeet/utils/grpc-middlewares/cid"
	"github.com/gomeet/gomeet/utils/jwt"
)

type Fields map[string]interface{}

func appendJWTNfo(ctx context.Context, f Fields) Fields {
	attrs := map[string]string{
		"jti":  "_jti",
		"sub":  "_sub",
		"role": "_role",
	}
	for _, kk := range attrs {
		f[kk] = "unknown"
	}

	// jwt informations
	jwtClaims, ok := ctx.Value("jwt").(jwt.Claims)
	if ok {
		for k, kk := range attrs {
			if v, ok := jwtClaims[k].(string); ok {
				f[kk] = v
			}
		}
	}

	return f
}

func ctxFields(ctx context.Context, f Fields) logrus.Fields {
	lF := logrus.Fields{}

	f["_cid"] = grpc_cid.Get(ctx)
	f = appendJWTNfo(ctx, f)

	for k, v := range f {
		lF[k] = v
	}

	return lF
}

func Debug(ctx context.Context, msg string, f Fields) {
	grpc_logrus.
		Extract(ctx).
		WithFields(ctxFields(ctx, f)).
		Debug(msg)
}

func Info(ctx context.Context, msg string, f Fields) {
	grpc_logrus.
		Extract(ctx).
		WithFields(ctxFields(ctx, f)).
		Warn(msg)
}

func Warn(ctx context.Context, msg string, err error, f Fields) {
	grpc_logrus.
		Extract(ctx).
		WithError(err).
		WithFields(ctxFields(ctx, f)).
		Warn(msg)
}

func Error(ctx context.Context, msg string, err error, f Fields) {
	grpc_logrus.
		Extract(ctx).
		WithError(err).
		WithFields(ctxFields(ctx, f)).
		Error(msg)
}
