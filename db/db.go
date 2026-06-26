package db

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDB() (*gorm.DB, error) {
	if os.Getenv("GO_ENV") == "dev" {
		err := godotenv.Load()
		if err != nil {
			return nil, err
		}
	}

	sslmode := os.Getenv("DB_SSLMODE")
	if sslmode == "" {
		sslmode = "require"
	}
	dns := (&url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD")),
		Host:     net.JoinHostPort(os.Getenv("DB_HOST"), os.Getenv("DB_PORT")),
		Path:     os.Getenv("DB_NAME"),
		RawQuery: "sslmode=" + sslmode,
	}).String()

	db, err := gorm.Open(postgres.Open(dns), &gorm.Config{
		TranslateError: true,
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}
	fmt.Println("Connected")

	return db, nil
}

func CloseDB(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalln(err)
	}
	if err := sqlDB.Close(); err != nil {
		log.Fatalln(err)
	}
}
