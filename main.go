package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/url"
	"os"

	"github.com/matsukawakohei/todo-app-go/controller"
	"github.com/matsukawakohei/todo-app-go/db"
	"github.com/matsukawakohei/todo-app-go/middleware"
	"github.com/matsukawakohei/todo-app-go/repository"
	"github.com/matsukawakohei/todo-app-go/router"
	"github.com/matsukawakohei/todo-app-go/usecase"
)

func main() {
	slog.SetDefault(slog.New(middleware.NewContextSlogHandler(slog.NewJSONHandler(os.Stdout, nil))))
	db, err := db.NewDB()
	if err != nil {
		log.Fatalln(err)
	}
	frontURL := os.Getenv("FE_URL")
	if frontURL == "" {
		slog.Error("FE_URL is empty", "frontURL", frontURL)
		os.Exit(1)
	}
	u, err := url.Parse(frontURL)
	if err != nil {
		slog.Error("FE_URL is not URL format", "error", err)
		os.Exit(1)
	}
	if u.Scheme == "" || u.Host == "" || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		slog.Error("FE_URL invalid value", "value", frontURL)
		os.Exit(1)
	}

	userRepository := repository.NewUserRepository(db)
	userValidator := usecase.NewUserValidator()
	userUsecase := usecase.NewUserUsecase(userRepository, userValidator)
	userController := controller.NewUserController(userUsecase)

	e := router.NewRouter(userController, frontURL)

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}
	if err := e.Start(fmt.Sprintf(":%s", port)); err != nil {
		e.Logger.Error("failed to start server", "error", err)
		os.Exit(1)
	}
}
