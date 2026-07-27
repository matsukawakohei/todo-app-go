package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/matsukawakohei/todo-app-go/domain"
	"gorm.io/gorm"
)

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

func TestGetUserByID(t *testing.T) {
	tests := []struct {
		testName  string
		userID    domain.UserID
		seedUser  *userModel
		wantUser  *domain.User
		wantErr   bool
		wantErrIs error
	}{
		{
			testName: "success",
			userID:   1,
			seedUser: &userModel{ID: 1, Name: "test", Email: "test1@example.com"},
			wantUser: &domain.User{ID: 1, Name: "test", Email: "test1@example.com"},
			wantErr:  false,
		},
		{
			testName:  "notFound",
			userID:    1,
			wantErr:   true,
			wantErrIs: domain.ErrNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			runInTx(t, func(tx *gorm.DB) {
				if tc.seedUser != nil {
					if err := tx.Create(tc.seedUser).Error; err != nil {
						t.Fatalf("failed create user err: %v", err)
					}

					repo := NewUserRepository(tx)
					got, err := repo.GetUserByID(t.Context(), tc.userID)
					if tc.wantErr {
						if err == nil {
							t.Fatalf("want err is not nil; but got nil")
						}
						if !errors.Is(err, tc.wantErrIs) {
							t.Errorf("want err %v; but got %v", tc.wantErrIs, err)
						}
						return
					}

					if err != nil {
						t.Fatalf("want err is nill; but got %v", err)
					}
					if got.ID != tc.wantUser.ID {
						t.Errorf("want user.ID %v;but got %v", tc.wantUser.ID, got.ID)
					}
					if got.Name != tc.wantUser.Name {
						t.Errorf("want user.Name %v; but got %v", tc.wantUser.Name, got.Name)
					}
					if got.Email != tc.wantUser.Email {
						t.Errorf("want user.Email %v; but got %v", tc.wantUser.Email, got.Email)
					}
				}
			})
		})
	}
}
