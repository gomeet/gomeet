package cid

import (
	"context"

	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
)

const (
	correlationId = "correlation_id"
)

func Get(ctx context.Context) string {
	return metautils.ExtractIncoming(ctx).Get(correlationId)
}

func Set(ctx context.Context, cid string) context.Context {
	nmd := metautils.ExtractOutgoing(ctx)
	if cid != "" {
		nmd.Add(correlationId, cid)
		nCtx := nmd.ToOutgoing(ctx)
		return nCtx
	}

	return ctx
}
