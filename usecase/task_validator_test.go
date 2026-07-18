package usecase

import (
	"strings"
	"testing"
)

func TestCreateTaskValidate(t *testing.T) {
	tests := []struct {
		testName   string
		input      CreateTaskInput
		wantErr    bool
		wantErrMsg string
	}{
		{
			testName: "titleRequired",
			input: CreateTaskInput{
				Labels: []uint{},
			},
			wantErr:    true,
			wantErrMsg: "title is required",
		},
		{
			testName: "titleMinCharOk",
			input: CreateTaskInput{
				Title:  "あ",
				Labels: []uint{},
			},
			wantErr: false,
		},
		{
			testName: "titleMaxCharOk",
			input: CreateTaskInput{
				Title:  strings.Repeat("あ", 50),
				Labels: []uint{},
			},
			wantErr: false,
		},
		{
			testName: "titleMaxOverNG",
			input: CreateTaskInput{
				Title:  strings.Repeat("あ", 51),
				Labels: []uint{},
			},
			wantErr:    true,
			wantErrMsg: "limited max 50 char",
		},
		{
			testName: "labelsEmptyOK",
			input: CreateTaskInput{
				Title:  "あ",
				Labels: []uint{},
			},
			wantErr: false,
		},
		{
			testName: "labelLessMinNG",
			input: CreateTaskInput{
				Title:  "あ",
				Labels: []uint{0},
			},
			wantErr:    true,
			wantErrMsg: "label id must be greater than 0",
		},
		{
			testName: "labelMinOK",
			input: CreateTaskInput{
				Title:  "あ",
				Labels: []uint{1},
			},
			wantErr: false,
		},
	}

	v := NewTaskValidator()
	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			err := v.CreateTaskValidate(t.Context(), tc.input)

			if !tc.wantErr {
				if err != nil {
					t.Errorf("want err is nil; but got %v", err)
				}
				return
			}

			if err == nil {
				t.Fatalf("want err; but err is nil")
			}

			if !strings.Contains(err.Error(), tc.wantErrMsg) {
				t.Errorf("want err message %v; but got %v", tc.wantErrMsg, err)
			}
		})
	}
}

func TestUpdateTaskValidate(t *testing.T) {
	tests := []struct {
		testName   string
		input      UpdateTaskInput
		wantErr    bool
		wantErrMsg string
	}{
		{
			testName: "titleRequired",
			input: UpdateTaskInput{
				Completed: false,
				Labels:    []uint{},
			},
			wantErr:    true,
			wantErrMsg: "title is required",
		},
		{
			testName: "titleMinCharOk",
			input: UpdateTaskInput{
				Title:     "あ",
				Completed: false,
				Labels:    []uint{},
			},
			wantErr: false,
		},
		{
			testName: "titleMaxCharOk",
			input: UpdateTaskInput{
				Title:     strings.Repeat("あ", 50),
				Completed: false,
				Labels:    []uint{},
			},
			wantErr: false,
		},
		{
			testName: "titleMaxOverNG",
			input: UpdateTaskInput{
				Title:     strings.Repeat("あ", 51),
				Completed: false,
				Labels:    []uint{},
			},
			wantErr:    true,
			wantErrMsg: "limited max 50 char",
		},
		{
			testName: "labelsEmptyOK",
			input: UpdateTaskInput{
				Title:     "あ",
				Completed: false,
				Labels:    []uint{},
			},
			wantErr: false,
		},
		{
			testName: "labelLessMinNG",
			input: UpdateTaskInput{
				Title:     "あ",
				Completed: false,
				Labels:    []uint{0},
			},
			wantErr:    true,
			wantErrMsg: "label id must be greater than 0",
		},
		{
			testName: "labelMinOK",
			input: UpdateTaskInput{
				Title:     "あ",
				Completed: false,
				Labels:    []uint{1},
			},
			wantErr: false,
		},
	}

	v := NewTaskValidator()
	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			err := v.UpdateTaskValidate(t.Context(), tc.input)

			if !tc.wantErr {
				if err != nil {
					t.Errorf("want err is nil; but got %v", err)
				}
				return
			}

			if err == nil {
				t.Fatalf("want err; but err is nil")
			}

			if !strings.Contains(err.Error(), tc.wantErrMsg) {
				t.Errorf("want err message %v; but got %v", tc.wantErrMsg, err)
			}
		})
	}
}
