package repository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/matsukawakohei/todo-app-go/domain"
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

	if err := db.AutoMigrate(&userModel{}); err != nil {
		fmt.Fprintf(os.Stderr, "failed to migrate: %v\n", err)
		os.Exit(1)
	}

	testDB = db
	code := m.Run()

	_ = db.Migrator().DropTable(&userModel{})
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

func TestGetUserByEmail(t *testing.T) {
	tests := []struct {
		testName   string
		seed       bool
		seedUser   userModel
		queryEmail string
		wantErr    bool
		wantErrIs  error
		wantName   string
		wantEmail  string
	}{
		{
			testName:   "success",
			seed:       true,
			seedUser:   userModel{Name: "テスト", Email: "test@example.com", Password: "hashed"},
			queryEmail: "test@example.com",
			wantName:   "テスト",
			wantEmail:  "test@example.com",
		},
		{
			testName:   "notFound",
			seed:       false,
			queryEmail: "nobody@example.com",
			wantErr:    true,
			wantErrIs:  domain.ErrNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			runInTx(t, func(tx *gorm.DB) {
				if tc.seed {
					if err := tx.Create(&tc.seedUser).Error; err != nil {
						t.Fatalf("failed to seed user: %v", err)
					}
				}

				repo := NewUserRepository(tx)
				got, err := repo.GetUserByEmail(context.Background(), tc.queryEmail)

				if tc.wantErr {
					if err == nil {
						t.Errorf("expected error but got nil")
						return
					}
					if tc.wantErrIs != nil && !errors.Is(err, tc.wantErrIs) {
						t.Errorf("expected error %v; got %v", tc.wantErrIs, err)
					}
					return
				}
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if got.Name != tc.wantName {
					t.Errorf("expected Name %q; got %q", tc.wantName, got.Name)
				}
				if got.Email != tc.wantEmail {
					t.Errorf("expected Email %q; got %q", tc.wantEmail, got.Email)
				}
				if got.ID == 0 {
					t.Errorf("expected non-zero ID")
				}
				if got.CreatedAt.IsZero() {
					t.Errorf("expected non-zero CreatedAt")
				}
			})
		})
	}
}

func TestCreateUser(t *testing.T) {
	tests := []struct {
		testName  string
		seed      bool
		seedEmail string
		input     domain.CreateUser
		wantErr   bool
		wantErrIs error
	}{
		{
			testName: "success",
			input: domain.CreateUser{
				Name:     "テスト",
				Email:    "new@example.com",
				Password: "hashedpassword",
			},
		},
		{
			testName:  "duplicateEmail",
			seed:      true,
			seedEmail: "dup@example.com",
			input: domain.CreateUser{
				Name:     "別のユーザー",
				Email:    "dup@example.com",
				Password: "hashedpassword",
			},
			wantErr:   true,
			wantErrIs: domain.ErrDuplicatedKey,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			runInTx(t, func(tx *gorm.DB) {
				if tc.seed {
					seed := userModel{Name: "既存", Email: tc.seedEmail, Password: "x"}
					if err := tx.Create(&seed).Error; err != nil {
						t.Fatalf("failed to seed user: %v", err)
					}
				}

				repo := NewUserRepository(tx)
				got, err := repo.CreateUser(context.Background(), tc.input)

				if tc.wantErr {
					if err == nil {
						t.Errorf("expected error but got nil")
						return
					}
					if tc.wantErrIs != nil && !errors.Is(err, tc.wantErrIs) {
						t.Errorf("expected error %v; got %v", tc.wantErrIs, err)
					}
					return
				}
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if got.ID == 0 {
					t.Errorf("expected non-zero ID after create")
				}
				if got.Name != tc.input.Name {
					t.Errorf("expected Name %q; got %q", tc.input.Name, got.Name)
				}
				if got.Email != tc.input.Email {
					t.Errorf("expected Email %q; got %q", tc.input.Email, got.Email)
				}
				if got.CreatedAt.IsZero() {
					t.Errorf("expected non-zero CreatedAt")
				}

				var persisted userModel
				if err := tx.Where("email = ?", tc.input.Email).First(&persisted).Error; err != nil {
					t.Errorf("row not found after create: %v", err)
				}
			})
		})
	}
}
