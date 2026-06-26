package middleware

import (
	"context"

	"github.com/labstack/echo/v5"
)

type requestIDKey struct{}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

func RequestIDFromCtx(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey{}).(string); ok {
		return id
	}
	return ""
}

// RequestIDToContext はEchoのRequestIDミドルウェアが設定した
// X-Request-IDをcontext.Contextに格納するEchoミドルウェア。
// e.Use(echoMiddleware.RequestID()) の後に登録すること。
func RequestIDToContext() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			id := c.Response().Header().Get(echo.HeaderXRequestID)
			ctx := WithRequestID(c.Request().Context(), id)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}
