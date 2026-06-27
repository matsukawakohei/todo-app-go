package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/matsukawakohei/todo-app-go/controller"
	"github.com/matsukawakohei/todo-app-go/db"
	"github.com/matsukawakohei/todo-app-go/middleware"
	"github.com/matsukawakohei/todo-app-go/repository"
	"github.com/matsukawakohei/todo-app-go/router"
	"github.com/matsukawakohei/todo-app-go/usecase"
)

func main() {
	slog.SetDefault(slog.New(middleware.NewContextSlogHandler(slog.Default().Handler())))
	db, err := db.NewDB()
	if err != nil {
		log.Fatalln(err)
	}
	userRepository := repository.NewUserRepository(db)
	userValidator := usecase.NewUserValidator()
	userUsecase := usecase.NewUserUsecase(userRepository, userValidator)
	userController := controller.NewUserController(userUsecase)

	e := router.NewRouter(userController)
	
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}
	if err := e.Start(fmt.Sprintf(":%s", port)); err != nil {
		e.Logger.Error("failed to start server", "error", err)
		os.Exit(1)
	}
}