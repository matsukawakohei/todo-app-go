package repository

import (
	"fmt"
	"os"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var testDB *gorm.DB

func TestMain(m *testing.M) {
	dsn := "postgres://todo:todo@localhost:5432/todo_test?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		TranslateError: true,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to test DB: %v\n", err)
		os.Exit(1)
	}

	if err := db.AutoMigrate(&userModel{}, &labelModel{}); err != nil {
		fmt.Fprintf(os.Stderr, "failed to migrate: %v\n", err)
		os.Exit(1)
	}

	testDB = db
	code := m.Run()

	_ = db.Migrator().DropTable(&userModel{}, &labelModel{})
	sqlDB, _ := db.DB()
	_ = sqlDB.Close()

	os.Exit(code)
}

func runInTx(t *testing.T, fn func(tx *gorm.DB)) {
	t.Helper()
	tx := testDB.Begin()
	if tx.Error != nil {
		t.Fatalf("failed to begin transaction: %v", tx.Error)
	}
	defer tx.Rollback()
	fn(tx)
}
