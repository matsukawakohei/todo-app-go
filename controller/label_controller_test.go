package controller

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/matsukawakohei/todo-app-go/domain"
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
		wantLabels       []*usecase.Label
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
			wantLabels: []*usecase.Label{},
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
			wantLabels: []*usecase.Label{
				{ID: 1, Name: "label1"},
				{ID: 2, Name: "label2"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			c, rec := newTestContext(http.MethodGet, "")
			if tc.loggedIn {
				setLoginUser(c, 1)
			}
			lc := NewLabelController(&mockLabelUsecase{getAllByUserIDFn: tc.getAllByUserIDFn})
			if err := lc.GetAllLabels(c); err != nil {
				t.Fatalf("label controller GetAllLabels failed: %v", err)
			}

			if rec.Code != tc.wantStatus {
				t.Errorf("expected status %d; got %d", tc.wantStatus, rec.Code)
			}

			if tc.wantLabels != nil {
				var output usecase.LabelGetAllOutput
				if err := json.Unmarshal(rec.Body.Bytes(), &output); err != nil {
					t.Fatalf("failed to unmarshal response body: %v", err)
				}
				if len(output.Labels) != len(tc.wantLabels) {
					t.Fatalf("expected %d labels; got %d", len(tc.wantLabels), len(output.Labels))
				}
				for i, l := range output.Labels {
					if *l != *tc.wantLabels[i] {
						t.Errorf("expected label %+v; got %+v", *tc.wantLabels[i], *l)
					}
				}
			}
		})
	}
}

func TestCreateLabel(t *testing.T) {
	tests := []struct {
		testName   string
		loggedIn   bool
		body       string
		createFn   func(ctx context.Context, userID domain.UserID, input usecase.LabelCreateInput) (*usecase.Label, error)
		wantStatus int
		wantLabel  *usecase.Label
	}{
		{
			testName:   "getLoginUserFromContextError",
			loggedIn:   false,
			body:       `{"Name":"label1"}`,
			wantStatus: http.StatusInternalServerError,
		},
		{
			testName:   "bindError",
			loggedIn:   true,
			body:       `{invalid}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			testName: "validationError",
			loggedIn: true,
			body:     `{"Name":""}`,
			createFn: func(_ context.Context, _ domain.UserID, _ usecase.LabelCreateInput) (*usecase.Label, error) {
				return nil, errors.New("Name: name is required.")
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			testName: "limitExceeded",
			loggedIn: true,
			body:     `{"Name":"label1"}`,
			createFn: func(_ context.Context, _ domain.UserID, _ usecase.LabelCreateInput) (*usecase.Label, error) {
				return nil, domain.ErrLimitExceeded
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			testName: "userNotFound",
			loggedIn: true,
			body:     `{"Name":"label1"}`,
			createFn: func(_ context.Context, _ domain.UserID, _ usecase.LabelCreateInput) (*usecase.Label, error) {
				return nil, domain.ErrNotFound
			},
			wantStatus: http.StatusNotFound,
		},
		{
			testName: "duplicatedKey",
			loggedIn: true,
			body:     `{"Name":"label1"}`,
			createFn: func(_ context.Context, _ domain.UserID, _ usecase.LabelCreateInput) (*usecase.Label, error) {
				return nil, domain.ErrDuplicatedKey
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			testName: "usecaseErrUnexpected",
			loggedIn: true,
			body:     `{"Name":"label1"}`,
			createFn: func(_ context.Context, _ domain.UserID, _ usecase.LabelCreateInput) (*usecase.Label, error) {
				return nil, domain.ErrUnexpected
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			testName: "success",
			loggedIn: true,
			body:     `{"Name":"label1"}`,
			createFn: func(_ context.Context, _ domain.UserID, _ usecase.LabelCreateInput) (*usecase.Label, error) {
				return &usecase.Label{ID: 1, Name: "label1"}, nil
			},
			wantStatus: http.StatusCreated,
			wantLabel:  &usecase.Label{ID: 1, Name: "label1"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			c, rec := newTestContext(http.MethodPost, tc.body)
			if tc.loggedIn {
				setLoginUser(c, 1)
			}
			lc := NewLabelController(&mockLabelUsecase{createFn: tc.createFn})
			if err := lc.CreateLabel(c); err != nil {
				t.Fatalf("label controller CreateLabel failed: %v", err)
			}

			if rec.Code != tc.wantStatus {
				t.Errorf("expected status %d; got %d", tc.wantStatus, rec.Code)
			}

			if tc.wantLabel != nil {
				var output usecase.Label
				if err := json.Unmarshal(rec.Body.Bytes(), &output); err != nil {
					t.Fatalf("failed to unmarshal response body: %v", err)
				}
				if output != *tc.wantLabel {
					t.Errorf("expected label %+v; got %+v", *tc.wantLabel, output)
				}
			}
		})
	}
}

func TestDeleteLabel(t *testing.T) {
	tests := []struct {
		testName   string
		loggedIn   bool
		labelIDStr string
		deleteFn   func(ctx context.Context, userID domain.UserID, labelID domain.LabelID) error
		wantStatus int
	}{
		{
			testName:   "getLoginUserFromContextError",
			loggedIn:   false,
			labelIDStr: "1",
			wantStatus: http.StatusInternalServerError,
		},
		{
			testName:   "getLabelIDFromParamError",
			loggedIn:   true,
			labelIDStr: "abc",
			wantStatus: http.StatusBadRequest,
		},
		{
			testName:   "notFound",
			loggedIn:   true,
			labelIDStr: "1",
			deleteFn: func(_ context.Context, _ domain.UserID, _ domain.LabelID) error {
				return domain.ErrNotFound
			},
			wantStatus: http.StatusNotFound,
		},
		{
			testName:   "usecaseErrUnexpected",
			loggedIn:   true,
			labelIDStr: "1",
			deleteFn: func(_ context.Context, _ domain.UserID, _ domain.LabelID) error {
				return domain.ErrUnexpected
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			testName:   "success",
			loggedIn:   true,
			labelIDStr: "1",
			deleteFn: func(_ context.Context, _ domain.UserID, _ domain.LabelID) error {
				return nil
			},
			wantStatus: http.StatusNoContent,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			c, rec := newTestContext(http.MethodDelete, "")
			if tc.loggedIn {
				setLoginUser(c, 1)
			}
			c.SetPathValues(echo.PathValues{{Name: "id", Value: tc.labelIDStr}})

			lc := NewLabelController(&mockLabelUsecase{deleteFn: tc.deleteFn})
			if err := lc.DeleteLabel(c); err != nil {
				t.Fatalf("label controller DeleteLabel failed: %v", err)
			}

			if rec.Code != tc.wantStatus {
				t.Errorf("expected status %d; got %d", tc.wantStatus, rec.Code)
			}
		})
	}
}
