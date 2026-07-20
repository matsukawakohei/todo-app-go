package middleware

import (
	"log/slog"

	"github.com/labstack/echo/v5"
)

// AccessLogger は全ルートでリクエスト/レスポンスの内容をslogで出力するEchoミドルウェア。
// /api/v1/auth配下はリクエストボディに平文パスワードが含まれるため、ボディのダンプ自体を行わない。
func AccessLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			err := next(c)
			_, status := echo.ResolveResponseStatus(c.Response(), err)
			slog.InfoContext(c.Request().Context(), "http request",
				"method", c.Request().Method,
				"path", c.Request().URL.Path,
				"status", status,
				"IP", c.Request().RemoteAddr,
			)

			return err
		}
	}
}
