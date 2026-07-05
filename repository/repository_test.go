package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/url"
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
	dsn := os.Getenv("TEST_DB_URL")
	if dsn == "" {
		fmt.Fprintln(os.Stderr, "TEST_DB_URL environment variable is not set")
		return 1
	}

	if err := validateTestDBURL(dsn); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	driver, err := migratePostgres.WithInstance(db, &migratePostgres.Config{})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	migrator, err := migrate.NewWithDatabaseInstance("file://../migration", "postgres", driver)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer func() {
		if err := migrator.Down(); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		if err := db.Close(); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}()

	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		TranslateError: true,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer func() {
		sqlDB, err := testDB.DB()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		if err := sqlDB.Close(); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}()
	testDB = gormDB

	return m.Run()
}

func validateTestDBURL(dsn string) error {
	u, err := url.Parse(dsn)
	if err != nil {
		return fmt.Errorf("TEST_DB_URL is not a valid URL: %w", err)
	}

	if u.Scheme != "postgres" && u.Scheme != "postgresql" {
		return fmt.Errorf("TEST_DB_URL must use the postgres:// scheme, got %q", u.Scheme)
	}

	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		return fmt.Errorf("TEST_DB_URL host %q is not a valid host:port: %w", u.Host, err)
	}
	if host == "" || port == "" {
		return fmt.Errorf("TEST_DB_URL host %q must include both a host and a port", u.Host)
	}

	return nil
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
