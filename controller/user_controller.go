package controller

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/matsukawakohei/todo-app-go/usecase"
)

type UserController interface {
	SignUp(c *echo.Context) error
	Login(c *echo.Context) error
	Logout(c *echo.Context) error
}

type userController struct {
	uu usecase.UserUsecase
}

func NewUserController(uu usecase.UserUsecase) UserController {
	return &userController{uu}
}

func (uc *userController) SignUp(c *echo.Context) error {
	input := usecase.SignUpInput{}
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	ctx := c.Request().Context()
	output, err := uc.uu.SignUp(ctx, input)
	if err != nil {
		return buildResponseError(c, err)
	}

	return c.JSON(http.StatusCreated, output)
}

func (uc *userController) Login(c *echo.Context) error {
	input := usecase.LoginInput{}
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	ctx := c.Request().Context()
	token, err := uc.uu.Login(ctx, input)
	if err != nil {
		return buildResponseError(c, err)
	}

	cookie := new(http.Cookie)
	cookie.Name = "token"
	cookie.Value = token
	cookie.Expires = time.Now().Add(time.Hour * 24)
	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.Secure = CookieSecure()
	cookie.SameSite = http.SameSiteLaxMode
	c.SetCookie(cookie)

	return c.NoContent(http.StatusOK)
}

func (uc *userController) Logout(c *echo.Context) error {
	cookie := new(http.Cookie)
	cookie.Name = "token"
	cookie.Value = ""
	cookie.Expires = time.Now()
	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.Secure = CookieSecure()
	cookie.SameSite = http.SameSiteLaxMode
	cookie.MaxAge = -1
	c.SetCookie(cookie)

	return c.NoContent(http.StatusOK)
}
