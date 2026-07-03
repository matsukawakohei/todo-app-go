package router

import (
	"net/http"

	"github.com/labstack/echo/v5"
	echoMiddleware "github.com/labstack/echo/v5/middleware"
	"github.com/matsukawakohei/todo-app-go/controller"
	"github.com/matsukawakohei/todo-app-go/middleware"
)

func NewRouter(uc controller.UserController, frontURL string) *echo.Echo {
	e := echo.New()
	e.Use(echoMiddleware.RequestID())
	e.Use(middleware.RequestIDToContext())

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
		CookieSecure:   controller.CookieSecure(),
	}))

	api := e.Group("/api")
	v1 := api.Group("/v1")

	auth := v1.Group("/auth")
	auth.POST("/signup", uc.SignUp)
	auth.POST("/login", uc.Login)
	auth.POST("/logout", uc.Logout)

	return e
}
