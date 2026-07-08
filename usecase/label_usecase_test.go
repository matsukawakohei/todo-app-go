package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/matsukawakohei/todo-app-go/domain"
)

type mockLabelRepository struct {
	getAllByUserIDFn func(ctx context.Context, userID domain.UserID) ([]*domain.Label, error)
	createLabelFn    func(ctx context.Context, userID domain.UserID, label domain.CreateLabel) (*domain.Label, error)
	deleteLabelFn    func(ctx context.Context, userID domain.UserID, labelID domain.LabelID) error
}

func (m *mockLabelRepository) GetAllByUserID(ctx context.Context, userID domain.UserID) ([]*domain.Label, error) {
	return m.getAllByUserIDFn(ctx, userID)
}

func (m *mockLabelRepository) CreateLabel(ctx context.Context, userID domain.UserID, label domain.CreateLabel) (*domain.Label, error) {
	return m.createLabelFn(ctx, userID, label)
}

func (m *mockLabelRepository) DeleteLabel(ctx context.Context, userID domain.UserID, labelID domain.LabelID) error {
	return m.deleteLabelFn(ctx, userID, labelID)
}

type mockLabelValidator struct {
	createValidateFn func(ctx context.Context, input LabelCreateInput) error
}

func (m *mockLabelValidator) CreateValidate(ctx context.Context, input LabelCreateInput) error {
	return m.createValidateFn(ctx, input)
}

func TestLabelGetAllByUserID(t *testing.T) {
	tests := []struct {
		testName   string
		userID     domain.UserID
		repoLabels []*domain.Label
		repoErr    error
		wantErr    bool
		wantLen    int
	}{
		{
			testName: "success",
			userID:   domain.UserID(1),
			repoLabels: []*domain.Label{
				{ID: domain.LabelID(1), Name: "label1"},
				{ID: domain.LabelID(2), Name: "label2"},
			},
			wantLen: 2,
		},
		{
			testName:   "empty",
			userID:     domain.UserID(1),
			repoLabels: []*domain.Label{},
			wantLen:    0,
		},
		{
			testName: "repositoryError",
			userID:   domain.UserID(1),
			repoErr:  domain.ErrUnexpected,
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			var capturedUserID domain.UserID
			mockRepo := &mockLabelRepository{
				getAllByUserIDFn: func(ctx context.Context, userID domain.UserID) ([]*domain.Label, error) {
					capturedUserID = userID
					return tc.repoLabels, tc.repoErr
				},
			}

			mockValidator := &mockLabelValidator{
				createValidateFn: func(ctx context.Context, input LabelCreateInput) error {
					return nil
				},
			}

			lu := NewLabelUsecase(mockRepo, mockValidator)
			output, err := lu.GetAllByUserID(t.Context(), tc.userID)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("expect err is not nil; but got nil")
				}
				if !errors.Is(err, tc.repoErr) {
					t.Errorf("expect err is %v; but got %v", tc.repoErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if capturedUserID != tc.userID {
				t.Errorf("expect userID passed to repository is %v; but got %v", tc.userID, capturedUserID)
			}
			if output.labels == nil {
				t.Errorf("expect labels is an empty slice, not nil, when there are no labels")
			}
			if len(output.labels) != tc.wantLen {
				t.Fatalf("expect %d labels; but got %d", tc.wantLen, len(output.labels))
			}
			for i, l := range output.labels {
				if l.ID != uint(tc.repoLabels[i].ID) {
					t.Errorf("expect label[%d].ID is %d; but got %d", i, tc.repoLabels[i].ID, l.ID)
				}
				if l.Name != tc.repoLabels[i].Name {
					t.Errorf("expect label[%d].Name is %s; but got %s", i, tc.repoLabels[i].Name, l.Name)
				}
			}
		})
	}
}

func TestLabelCreate(t *testing.T) {
	tests := []struct {
		testName    string
		userID      domain.UserID
		input       LabelCreateInput
		validateErr error
		repoLabel   *domain.Label
		repoErr     error
		wantErr     bool
		wantErrIs   error
		wantCalled  bool
	}{
		{
			testName:   "success",
			userID:     domain.UserID(1),
			input:      LabelCreateInput{Name: "label1"},
			repoLabel:  &domain.Label{ID: domain.LabelID(1), Name: "label1"},
			wantCalled: true,
		},
		{
			testName:    "validationError",
			userID:      domain.UserID(1),
			input:       LabelCreateInput{Name: ""},
			validateErr: errors.New("Name: name is required."),
			wantErr:     true,
		},
		{
			testName:   "repositoryUnexpectedError",
			userID:     domain.UserID(1),
			input:      LabelCreateInput{Name: "label1"},
			repoErr:    domain.ErrUnexpected,
			wantErr:    true,
			wantErrIs:  domain.ErrUnexpected,
			wantCalled: true,
		},
		{
			testName:   "repositoryLimitExceeded",
			userID:     domain.UserID(1),
			input:      LabelCreateInput{Name: "label1"},
			repoErr:    domain.ErrLimitExceeded,
			wantErr:    true,
			wantErrIs:  domain.ErrLimitExceeded,
			wantCalled: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			var called bool
			var capturedUserID domain.UserID
			var capturedInput domain.CreateLabel
			mockRepo := &mockLabelRepository{
				createLabelFn: func(ctx context.Context, userID domain.UserID, label domain.CreateLabel) (*domain.Label, error) {
					called = true
					capturedUserID = userID
					capturedInput = label
					return tc.repoLabel, tc.repoErr
				},
			}

			mockValidator := &mockLabelValidator{
				createValidateFn: func(ctx context.Context, input LabelCreateInput) error {
					return tc.validateErr
				},
			}

			lu := NewLabelUsecase(mockRepo, mockValidator)
			label, err := lu.Create(t.Context(), tc.userID, tc.input)

			if called != tc.wantCalled {
				t.Errorf("expect repository called is %v; but got %v", tc.wantCalled, called)
			}

			if tc.wantErr {
				if err == nil {
					t.Fatalf("expect err is not nil; but got nil")
				}
				if tc.wantErrIs != nil && !errors.Is(err, tc.wantErrIs) {
					t.Errorf("expect err is %v; but got %v", tc.wantErrIs, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if capturedUserID != tc.userID {
				t.Errorf("expect userID passed to repository is %v; but got %v", tc.userID, capturedUserID)
			}
			if capturedInput.Name != tc.input.Name {
				t.Errorf("expect CreateLabel.Name is %s; but got %s", tc.input.Name, capturedInput.Name)
			}
			if label.ID != uint(tc.repoLabel.ID) {
				t.Errorf("expect label.ID is %d; but got %d", tc.repoLabel.ID, label.ID)
			}
			if label.Name != tc.repoLabel.Name {
				t.Errorf("expect label.Name is %s; but got %s", tc.repoLabel.Name, label.Name)
			}
		})
	}
}

func TestLabelDelete(t *testing.T) {
	tests := []struct {
		testName string
		userID   domain.UserID
		labelID  domain.LabelID
		repoErr  error
		wantErr  bool
	}{
		{
			testName: "success",
			userID:   domain.UserID(1),
			labelID:  domain.LabelID(1),
		},
		{
			testName: "notFound",
			userID:   domain.UserID(1),
			labelID:  domain.LabelID(999),
			repoErr:  domain.ErrNotFound,
			wantErr:  true,
		},
		{
			testName: "repositoryUnexpectedError",
			userID:   domain.UserID(1),
			labelID:  domain.LabelID(1),
			repoErr:  domain.ErrUnexpected,
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			var capturedUserID domain.UserID
			var capturedLabelID domain.LabelID
			mockRepo := &mockLabelRepository{
				deleteLabelFn: func(ctx context.Context, userID domain.UserID, labelID domain.LabelID) error {
					capturedUserID = userID
					capturedLabelID = labelID
					return tc.repoErr
				},
			}

			mockValidator := &mockLabelValidator{
				createValidateFn: func(ctx context.Context, input LabelCreateInput) error {
					return nil
				},
			}

			lu := NewLabelUsecase(mockRepo, mockValidator)
			err := lu.Delete(t.Context(), tc.userID, tc.labelID)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("expect err is not nil; but got nil")
				}
				if !errors.Is(err, tc.repoErr) {
					t.Errorf("expect err is %v; but got %v", tc.repoErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if capturedUserID != tc.userID {
				t.Errorf("expect userID passed to repository is %v; but got %v", tc.userID, capturedUserID)
			}
			if capturedLabelID != tc.labelID {
				t.Errorf("expect labelID passed to repository is %v; but got %v", tc.labelID, capturedLabelID)
			}
		})
	}
}
