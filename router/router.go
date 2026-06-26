package router

import (
	"github.com/labstack/echo/v5"
	echoMiddleware "github.com/labstack/echo/v5/middleware"
	"github.com/matsukawakohei/todo-app-go/controller"
	"github.com/matsukawakohei/todo-app-go/middleware"
)

func NewRouter(uc controller.UserController) *echo.Echo {
	e := echo.New()
	e.Use(echoMiddleware.RequestID())
	e.Use(middleware.RequestIDToContext())

	// TODO CSRF

	// TODO CORS

	api := e.Group("/api")
	v1 := api.Group("/v1")

	auth := v1.Group("/auth")
	auth.POST("/signup", uc.SignUp)
	auth.POST("/login", uc.Login)
	auth.POST("/logout", uc.Logout)

	return e
}