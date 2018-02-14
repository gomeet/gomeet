package context

import (
	"errors"
	"fmt"

	"github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	grpc_cid "github.com/gomeet/gomeet/utils/grpc-middlewares/cid"
	"github.com/gomeet/gomeet/utils/jwt"
	"github.com/gomeet/gomeet/utils/log"
)

func ParseJWTFromContext(ctx context.Context, jwtSecret string) (context.Context, error) {
	if jwtSecret == "" {
		return nil, errors.New("No jwt secret")
	}

	// check the bearer token
	bearerToken, err := grpc_auth.AuthFromMD(ctx, "bearer")
	if err != nil {
		log.Warn(ctx, "failed to read JWT bearer token", err, log.Fields{})
		return nil, err
	}

	claims, err := jwt.Parse(jwtSecret, bearerToken)
	if err != nil {
		log.Error(ctx, "invalid auth token", err, log.Fields{})
		return nil, status.Errorf(codes.Unauthenticated, "invalid auth token")
	}

	// inject token claims into the context
	newCtx := context.WithValue(ctx, "jwt", claims)

	return newCtx, nil
}

// AuthContextFromJWT : add authorization header from jwt to ctx
func AuthContextFromJWT(ctx context.Context, jwt string) context.Context {
	nmd := metautils.ExtractOutgoing(ctx)
	if jwt != "" {
		nmd.Add("Authorization", fmt.Sprintf("Bearer %s", jwt))
		nCtx := nmd.ToOutgoing(ctx)
		return nCtx
	}

	return ctx
}

func NewSubServiceContext(ctx context.Context) context.Context {
	// jwt transfer
	jwt, err := grpc_auth.AuthFromMD(ctx, "bearer")
	if err != nil {
		jwt = ""
	}

	return NewSubServiceContextWithJWT(ctx, jwt)
}

func NewSubServiceContextWithJWT(ctx context.Context, jwt string) context.Context {
	nCtx := context.Background()
	nCtx = AuthContextFromJWT(nCtx, jwt)
	// correlationId transfer
	cid := grpc_cid.Get(ctx)
	if cid != "" {
		nCtx = grpc_cid.Set(nCtx, cid)
	}
	return nCtx
}
