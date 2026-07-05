package repository

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	migratePostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var testDB *gorm.DB

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	dsn := "postgres://todo:todo@localhost:5432/todo_test?sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	driver, err := migratePostgres.WithInstance(db, &migratePostgres.Config{})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	migrator, err := migrate.NewWithDatabaseInstance("file://../migration", "postgres", driver)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer func() {
		migrator.Down()
		db.Close()
	}()

	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		TranslateError: true,
	})
	defer func() {
		sqlDB, _ := testDB.DB()
		sqlDB.Close()
	}()
	testDB = gormDB

	return m.Run()
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
