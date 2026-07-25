package handler

import (
	"context"
	"net/http"
	"testing"

	"github.com/matsukawakohei/todo-app-go/domain"
	"github.com/matsukawakohei/todo-app-go/internal/api"
	"github.com/matsukawakohei/todo-app-go/usecase"
)

type mockLabelUsecase struct {
	getAllByUserIDFn func(ctx context.Context, userID domain.UserID) (*usecase.LabelGetAllOutput, error)
	createFn         func(ctx context.Context, userID domain.UserID, input usecase.LabelCreateInput) (*usecase.Label, error)
	deleteFn         func(ctx context.Context, userID domain.UserID, labelID domain.LabelID) error
}

func (m *mockLabelUsecase) GetAllByUserID(ctx context.Context, userID domain.UserID) (*usecase.LabelGetAllOutput, error) {
	return m.getAllByUserIDFn(ctx, userID)
}

func (m *mockLabelUsecase) Create(ctx context.Context, userID domain.UserID, input usecase.LabelCreateInput) (*usecase.Label, error) {
	return m.createFn(ctx, userID, input)
}

func (m *mockLabelUsecase) Delete(ctx context.Context, userID domain.UserID, labelID domain.LabelID) error {
	return m.deleteFn(ctx, userID, labelID)
}

func TestGetAllLabels(t *testing.T) {
	tests := []struct {
		testName         string
		loggedIn         bool
		getAllByUserIDFn func(ctx context.Context, userID domain.UserID) (*usecase.LabelGetAllOutput, error)
		wantStatus       int
		wantLabels       []api.Label
	}{
		{
			testName:   "getLoginUserFromContextError",
			loggedIn:   false,
			wantStatus: http.StatusInternalServerError,
		},
		{
			testName: "usecaseErrUnexpected",
			loggedIn: true,
			getAllByUserIDFn: func(_ context.Context, _ domain.UserID) (*usecase.LabelGetAllOutput, error) {
				return nil, domain.ErrUnexpected
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			testName: "empty",
			loggedIn: true,
			getAllByUserIDFn: func(_ context.Context, _ domain.UserID) (*usecase.LabelGetAllOutput, error) {
				return &usecase.LabelGetAllOutput{Labels: []*usecase.Label{}}, nil
			},
			wantStatus: http.StatusOK,
			wantLabels: []api.Label{},
		},
		{
			testName: "success",
			loggedIn: true,
			getAllByUserIDFn: func(_ context.Context, _ domain.UserID) (*usecase.LabelGetAllOutput, error) {
				return &usecase.LabelGetAllOutput{Labels: []*usecase.Label{
					{ID: 1, Name: "label1"},
					{ID: 2, Name: "label2"},
				}}, nil
			},
			wantStatus: http.StatusOK,
			wantLabels: []api.Label{
				{Id: 1, Name: "label1"},
				{Id: 2, Name: "label2"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			ctx := context.Background()
			if tc.loggedIn {
				ctx = contextWithUser(1)
			}
			lh := NewLabelHandler(&mockLabelUsecase{getAllByUserIDFn: tc.getAllByUserIDFn})
			response, err := lh.GetLabels(ctx, api.GetLabelsRequestObject{})
			if err != nil {
				t.Fatalf("label handler GetLabels failed: %v", err)
			}

			switch tc.wantStatus {
			case http.StatusOK:
				got, ok := response.(api.GetLabels200JSONResponse)
				if !ok {
					t.Fatalf("expected api.GetLabels200JSONResponse; got %T", response)
				}
				if len(got.Data) != len(tc.wantLabels) {
					t.Fatalf("expected %d labels; got %d", len(tc.wantLabels), len(got.Data))
				}
				for i, l := range got.Data {
					if l != tc.wantLabels[i] {
						t.Errorf("expected label %+v; got %+v", tc.wantLabels[i], l)
					}
				}
			case http.StatusInternalServerError:
				if _, ok := response.(api.GetLabels500JSONResponse); !ok {
					t.Fatalf("expected api.GetLabels500JSONResponse; got %T", response)
				}
			}
		})
	}
}

func TestCreateLabel(t *testing.T) {
	tests := []struct {
		testName   string
		loggedIn   bool
		name       string
		createFn   func(ctx context.Context, userID domain.UserID, input usecase.LabelCreateInput) (*usecase.Label, error)
		wantStatus int
		wantLabel  *api.Label
	}{
		{
			testName:   "getLoginUserFromContextError",
			loggedIn:   false,
			name:       "label1",
			wantStatus: http.StatusInternalServerError,
		},
		{
			testName: "validationError",
			loggedIn: true,
			name:     "",
			createFn: func(_ context.Context, _ domain.UserID, _ usecase.LabelCreateInput) (*usecase.Label, error) {
				return nil, domain.ErrBadRequest
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			testName: "limitExceeded",
			loggedIn: true,
			name:     "label1",
			createFn: func(_ context.Context, _ domain.UserID, _ usecase.LabelCreateInput) (*usecase.Label, error) {
				return nil, domain.ErrLimitExceeded
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			testName: "userNotFound",
			loggedIn: true,
			name:     "label1",
			createFn: func(_ context.Context, _ domain.UserID, _ usecase.LabelCreateInput) (*usecase.Label, error) {
				return nil, domain.ErrNotFound
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			testName: "duplicatedKey",
			loggedIn: true,
			name:     "label1",
			createFn: func(_ context.Context, _ domain.UserID, _ usecase.LabelCreateInput) (*usecase.Label, error) {
				return nil, domain.ErrDuplicatedKey
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			testName: "usecaseErrUnexpected",
			loggedIn: true,
			name:     "label1",
			createFn: func(_ context.Context, _ domain.UserID, _ usecase.LabelCreateInput) (*usecase.Label, error) {
				return nil, domain.ErrUnexpected
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			testName: "success",
			loggedIn: true,
			name:     "label1",
			createFn: func(_ context.Context, _ domain.UserID, _ usecase.LabelCreateInput) (*usecase.Label, error) {
				return &usecase.Label{ID: 1, Name: "label1"}, nil
			},
			wantStatus: http.StatusCreated,
			wantLabel:  &api.Label{Id: 1, Name: "label1"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			ctx := context.Background()
			if tc.loggedIn {
				ctx = contextWithUser(1)
			}
			lh := NewLabelHandler(&mockLabelUsecase{createFn: tc.createFn})
			request := api.CreateLabelRequestObject{Body: &api.LabelInput{Name: tc.name}}
			response, err := lh.CreateLabel(ctx, request)
			if err != nil {
				t.Fatalf("label handler CreateLabel failed: %v", err)
			}

			switch tc.wantStatus {
			case http.StatusCreated:
				got, ok := response.(api.CreateLabel201JSONResponse)
				if !ok {
					t.Fatalf("expected api.CreateLabel201JSONResponse; got %T", response)
				}
				if got.Data != *tc.wantLabel {
					t.Errorf("expected label %+v; got %+v", *tc.wantLabel, got.Data)
				}
			case http.StatusBadRequest:
				if _, ok := response.(api.CreateLabel400JSONResponse); !ok {
					t.Fatalf("expected api.CreateLabel400JSONResponse; got %T", response)
				}
			case http.StatusInternalServerError:
				if _, ok := response.(api.CreateLabel500JSONResponse); !ok {
					t.Fatalf("expected api.CreateLabel500JSONResponse; got %T", response)
				}
			}
		})
	}
}

func TestDeleteLabel(t *testing.T) {
	tests := []struct {
		testName   string
		loggedIn   bool
		labelID    int
		deleteFn   func(ctx context.Context, userID domain.UserID, labelID domain.LabelID) error
		wantStatus int
	}{
		{
			testName:   "getLoginUserFromContextError",
			loggedIn:   false,
			labelID:    1,
			wantStatus: http.StatusInternalServerError,
		},
		{
			testName:   "negativeLabelID",
			loggedIn:   true,
			labelID:    -1,
			wantStatus: http.StatusBadRequest,
		},
		{
			testName: "notFound",
			loggedIn: true,
			labelID:  1,
			deleteFn: func(_ context.Context, _ domain.UserID, _ domain.LabelID) error {
				return domain.ErrNotFound
			},
			wantStatus: http.StatusNotFound,
		},
		{
			testName: "usecaseErrUnexpected",
			loggedIn: true,
			labelID:  1,
			deleteFn: func(_ context.Context, _ domain.UserID, _ domain.LabelID) error {
				return domain.ErrUnexpected
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			testName: "success",
			loggedIn: true,
			labelID:  1,
			deleteFn: func(_ context.Context, _ domain.UserID, _ domain.LabelID) error {
				return nil
			},
			wantStatus: http.StatusNoContent,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			ctx := context.Background()
			if tc.loggedIn {
				ctx = contextWithUser(1)
			}

			lh := NewLabelHandler(&mockLabelUsecase{deleteFn: tc.deleteFn})
			response, err := lh.DeleteLabel(ctx, api.DeleteLabelRequestObject{Id: tc.labelID})
			if err != nil {
				t.Fatalf("label handler DeleteLabel failed: %v", err)
			}

			switch tc.wantStatus {
			case http.StatusNoContent:
				if _, ok := response.(api.DeleteLabel204Response); !ok {
					t.Fatalf("expected api.DeleteLabel204Response; got %T", response)
				}
			case http.StatusBadRequest:
				if _, ok := response.(api.DeleteLabel400JSONResponse); !ok {
					t.Fatalf("expected api.DeleteLabel400JSONResponse; got %T", response)
				}
			case http.StatusNotFound:
				if _, ok := response.(api.DeleteLabel404JSONResponse); !ok {
					t.Fatalf("expected api.DeleteLabel404JSONResponse; got %T", response)
				}
			case http.StatusInternalServerError:
				if _, ok := response.(api.DeleteLabel500JSONResponse); !ok {
					t.Fatalf("expected api.DeleteLabel500JSONResponse; got %T", response)
				}
			}
		})
	}
}
