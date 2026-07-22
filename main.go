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
	gormDB, err := db.NewDB()
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
	if u.Scheme != "http" && u.Scheme != "https" {
		slog.Error("FE_URL invalid scheme", "value", frontURL)
		os.Exit(1)
	}
	if u.Host == "" || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		slog.Error("FE_URL invalid value", "value", frontURL)
		os.Exit(1)
	}
	secret := os.Getenv("SECRET")
	if secret == "" {
		slog.Error("SECRET is empty")
		os.Exit(1)
	}
	if len(secret) < 32 {
		slog.Error("SECRET is too short")
		os.Exit(1)
	}

	if err := db.Migrate(); err != nil {
		log.Fatalln(err)
	}

	userRepository := repository.NewUserRepository(gormDB)
	userValidator := usecase.NewUserValidator()
	userUsecase := usecase.NewUserUsecase(userRepository, userValidator, secret)
	userController := controller.NewUserController(userUsecase)

	labelRepository := repository.NewLabelRepository(gormDB)
	labelValidator := usecase.NewLabelValidator()
	labelUsecase := usecase.NewLabelUsecase(labelRepository, labelValidator)
	labelController := controller.NewLabelController(labelUsecase)

	taskRepository := repository.NewTaskRepository(gormDB)
	taskValidator := usecase.NewTaskValidator()
	taskUsecase := usecase.NewTaskUsecase(taskRepository, taskValidator)
	taskController := controller.NewTaskController(taskUsecase)

	e := router.NewRouter(userController, labelController, taskController, frontURL, secret)

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}
	if err := e.Start(fmt.Sprintf(":%s", port)); err != nil {
		e.Logger.Error("failed to start server", "error", err)
		os.Exit(1)
	}
}
