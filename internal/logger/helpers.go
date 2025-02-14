package logger

import (
	"context"

	"github.com/rs/zerolog"
)

func AddField(ctx context.Context, key string, val any) {
	zerolog.Ctx(ctx).UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Logger().With().Any(key, val)
	})
}

func AddError(ctx context.Context, err error) {
	zerolog.Ctx(ctx).UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Logger().With().Err(err)
	})
}
