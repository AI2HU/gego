package api

import (
	"context"
	"time"
)

const etcdRequestTimeout = 5 * time.Second

func etcdRequestContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); ok {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, etcdRequestTimeout)
}
