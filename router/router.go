package router

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v5"
	"github.com/labstack/echo/v5"
	echoMiddleware "github.com/labstack/echo/v5/middleware"
	"github.com/matsukawakohei/todo-app-go/internal/api"
	"github.com/matsukawakohei/todo-app-go/internal/handler"
	"github.com/matsukawakohei/todo-app-go/middleware"
)

func AuthMiddleware(secret string) echo.MiddlewareFunc {
	return echojwt.WithConfig(echojwt.Config{
		SigningKey:  []byte(secret),
		TokenLookup: "cookie:token",
		SuccessHandler: func(c *echo.Context) error {
			token, ok := c.Get("user").(*jwt.Token)
			if !ok {
				return nil
			}
			ctx := handler.ContextWithUserToken(c.Request().Context(), token)
			c.SetRequest(c.Request().WithContext(ctx))
			return nil
		},
	})
}

func NewRouter(ssi api.StrictServerInterface, frontURL string, secret string) *echo.Echo {
	e := echo.New()
	e.Use(echoMiddleware.RequestID())
	e.Use(middleware.RequestIDToContext())
	e.Use(middleware.AccessLogger())

	e.Use(echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
		AllowOrigins: []string{frontURL},
		AllowHeaders: []string{
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderXCSRFToken,
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowCredentials: true,
	}))

	e.Use(echoMiddleware.CSRFWithConfig(echoMiddleware.CSRFConfig{
		TrustedOrigins: []string{frontURL},
		CookieSameSite: http.SameSiteLaxMode,
		CookieSecure:   handler.CookieSecure(),
	}))

	authMiddleware := AuthMiddleware(secret)
	options := api.RegisterHandlersOptions{
		OperationMiddlewares: map[string][]echo.MiddlewareFunc{
			"getLabels": {
				authMiddleware,
			},
			"createLabel": {
				authMiddleware,
			},
			"deleteLabel": {
				authMiddleware,
			},
			"getTasks": {
				authMiddleware,
			},
			"createTask": {
				authMiddleware,
			},
			"updateTask": {
				authMiddleware,
			},
			"deleteTask": {
				authMiddleware,
			},
		},
	}

	si := api.NewStrictHandler(ssi, nil)
	api.RegisterHandlersWithOptions(e, si, options)

	return e
}
