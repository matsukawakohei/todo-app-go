package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/matsukawakohei/todo-app-go/domain"
	"gorm.io/gorm"
)

func TestGetAllByUserID(t *testing.T) {
	tests := []struct {
		testName    string
		seed        bool
		seedUsers   []userModel
		seedLabels  []labelModel
		queryUserID uint
		wantLabels  []domain.Label
	}{
		{
			testName:    "NotingData",
			seed:        false,
			queryUserID: 1,
			seedUsers:   []userModel{},
			wantLabels:  []domain.Label{},
		},
		{
			testName: "OtherUserLabel",
			seed:     true,
			seedUsers: []userModel{
				userModel{ID: 999, Name: "テスト", Email: "test@example.com"},
			},
			seedLabels: []labelModel{
				labelModel{Name: "テスト", UserID: 999},
			},
			queryUserID: 1,
			wantLabels:  []domain.Label{},
		},
		{
			testName: "MyLabel",
			seed:     true,
			seedUsers: []userModel{
				userModel{ID: 1, Name: "テスト", Email: "test@example.com"},
				userModel{ID: 999, Name: "テスト", Email: "test2@example.com"},
			},
			seedLabels: []labelModel{
				labelModel{ID: 1, Name: "テスト1", UserID: 1},
				labelModel{ID: 2, Name: "テスト2", UserID: 1},
				labelModel{ID: 3, Name: "テスト", UserID: 999},
			},
			queryUserID: 1,
			wantLabels: []domain.Label{
				domain.Label{ID: 2, Name: "テスト2"},
				domain.Label{ID: 1, Name: "テスト1"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			runInTx(t, func(tx *gorm.DB) {
				if len(tc.seedUsers) > 0 {
					if result := tx.Create(&tc.seedUsers); result.Error != nil {
						t.Fatalf("failed to seed users: %v", result.Error)
					}
				}
				if tc.seed && tc.seedLabels != nil && len(tc.seedLabels) > 0 {
					if result := tx.Create(&tc.seedLabels); result.Error != nil {
						t.Fatalf("failed to seed labels: %v", result.Error)
					}
				}

				repo := NewLabelRepository(tx)
				got, err := repo.GetAllByUserID(context.Background(), domain.UserID(tc.queryUserID))
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if len(tc.wantLabels) != len(got) {
					t.Fatalf("want len %v: but got %v", len(tc.wantLabels), len(got))
				}

				if len(tc.wantLabels) > 0 {
					for i := 0; i < len(tc.wantLabels); i++ {
						wantLabel := tc.wantLabels[i]
						gotLabel := got[i]

						if wantLabel.ID != gotLabel.ID || wantLabel.Name != gotLabel.Name {
							t.Errorf("want label %+v; but got label %+v", wantLabel, gotLabel)
						}
					}
				}
			})
		})
	}
}

func TestCreateLabel(t *testing.T) {
	tests := []struct {
		testName  string
		seedUser  userModel
		seedLabel *labelModel
		labelName string
		userID    uint
		wantErr   bool
		wantErrIs error
	}{
		{
			testName:  "success",
			seedUser:  userModel{ID: 1, Name: "テスト", Email: "test@example.com"},
			labelName: "テスト",
			userID:    1,
			wantErr:   false,
		},
		{
			testName:  "duplicateKey",
			seedUser:  userModel{ID: 1, Name: "テスト", Email: "test@example.com"},
			seedLabel: &labelModel{Name: "テスト", UserID: 1},
			labelName: "テスト",
			userID:    1,
			wantErr:   true,
			wantErrIs: domain.ErrDuplicatedKey,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			runInTx(t, func(tx *gorm.DB) {
				if result := tx.Create(&tc.seedUser); result.Error != nil {
					t.Fatalf("failed to seed user: %v", result.Error)
				}
				if tc.seedLabel != nil {
					if result := tx.Create(tc.seedLabel); result.Error != nil {
						t.Fatalf("failed to seed label: %v", result.Error)
					}
				}

				repo := NewLabelRepository(tx)
				input := domain.CreateLabel{
					Name: tc.labelName,
				}
				got, err := repo.CreateLabel(context.Background(), domain.UserID(tc.userID), input)
				if tc.wantErr {
					if err == nil {
						t.Fatalf("want err; but got nil")
					}
					if !errors.Is(err, tc.wantErrIs) {
						t.Errorf("want err %v; but got err %v", tc.wantErrIs, err)
					}
				} else {
					if err != nil {
						t.Fatalf("want error is nil; but got %v", err)
					}
					if got.Name != tc.labelName {
						t.Errorf("want label name %v; but got %v", tc.labelName, got.Name)
					}
				}
			})
		})
	}
}

func TestDeleteLabel(t *testing.T) {
	tests := []struct {
		testName  string
		seedUser  userModel
		seedLabel labelModel
		userID    uint
		labelID   uint
		wantErr   bool
		wantErrIs error
	}{
		{
			testName:  "success",
			seedUser:  userModel{ID: 1, Name: "テスト", Email: "test@example.com"},
			seedLabel: labelModel{ID: 1, Name: "テスト", UserID: 1},
			userID:    1,
			labelID:   1,
			wantErr:   false,
		},
		{
			testName:  "NotFound",
			seedUser:  userModel{ID: 1, Name: "テスト", Email: "test@example.com"},
			seedLabel: labelModel{ID: 1, Name: "テスト", UserID: 1},
			userID:    1,
			labelID:   2,
			wantErr:   true,
			wantErrIs: domain.ErrNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			runInTx(t, func(tx *gorm.DB) {
				if result := tx.Create(&tc.seedUser); result.Error != nil {
					t.Fatalf("failed to seed user: %v", result.Error)
				}
				if result := tx.Create(&tc.seedLabel); result.Error != nil {
					t.Fatalf("failed to seed label: %v", result.Error)
				}

				repo := NewLabelRepository(tx)
				ctx := context.Background()
				userID := domain.UserID(tc.userID)
				err := repo.DeleteLabel(ctx, userID, domain.LabelID(tc.labelID))

				if tc.wantErr {
					if err == nil {
						t.Fatalf("want err; but got nil")
					}
					if !errors.Is(err, tc.wantErrIs) {
						t.Errorf("want err %v; but got %v", tc.wantErrIs, err)
					}
				} else {
					if err != nil {
						t.Fatalf("DeleteLabel want err nil; but got %v", err)
					}

					labels, err := repo.GetAllByUserID(ctx, userID)
					if err != nil {
						t.Fatalf("GetAllByUserID want err nil; but got %v", err)
					}

					if len(labels) != 0 {
						t.Errorf("want labels len 0; but got %v", len(labels))
					}
				}
			})
		})
	}
}
