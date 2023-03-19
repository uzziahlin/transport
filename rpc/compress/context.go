package compress

import "context"

type Enabled struct {
}

func Context(ctx context.Context, typ Type) context.Context {
	return context.WithValue(ctx, Enabled{}, typ)
}

func EnableCompress(ctx context.Context) (Type, bool) {
	val := ctx.Value(Enabled{})

	if val == nil {
		return 0, false
	}

	typ, ok := val.(Type)

	return typ, ok
}
