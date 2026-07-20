package middleware

import (
	"log/slog"
	"strings"

	"github.com/labstack/echo/v5"
	echoMiddleware "github.com/labstack/echo/v5/middleware"
)

// AccessLogger は全ルートでリクエスト/レスポンスの内容をslogで出力するEchoミドルウェア。
// /api/v1/auth配下はリクエストボディに平文パスワードが含まれるため、ボディのダンプ自体を行わない。
func AccessLogger() echo.MiddlewareFunc {
	mw, err := echoMiddleware.BodyDumpConfig{
		Skipper: func(c *echo.Context) bool {
			return strings.HasPrefix(c.Request().URL.Path, "/api/v1/auth")
		},
		Handler: func(c *echo.Context, reqBody []byte, resBody []byte, err error) {
			_, status := echo.ResolveResponseStatus(c.Response(), err)
			slog.InfoContext(c.Request().Context(), "http request",
				"method", c.Request().Method,
				"path", c.Request().URL.Path,
				"status", status,
				"request_body", string(reqBody),
				"response_body", string(resBody),
			)
		},
	}.ToMiddleware()
	if err != nil {
		panic(err)
	}

	return mw
}
