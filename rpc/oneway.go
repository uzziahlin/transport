package rpc

import "context"

type OnewayKey struct{}

func OnewayContext(ctx context.Context) context.Context {
	vaLCtx := context.WithValue(ctx, OnewayKey{}, "true")
	return vaLCtx
}

func isOneway(ctx context.Context) bool {
	val := ctx.Value(OnewayKey{})
	bl, ok := val.(string)
	return ok && bl == "true"
}
