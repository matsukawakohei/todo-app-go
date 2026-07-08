package usecase

import (
	"context"

	"github.com/matsukawakohei/todo-app-go/domain"
)

type Label struct {
	ID uint
	Name string
}

func fromLabelEntity(label *domain.Label) *Label {
	return &Label{
		ID: uint(label.ID),
		Name: label.Name,
	}
}

type LabelGetAllOutput struct {
	labels []*Label
}

type LabelCreateInput struct {
	Name string
}

func (lci *LabelCreateInput) toEntityInput() domain.CreateLabel {
	return  domain.CreateLabel{Name: lci.Name}
}

type LabelUsecase interface {
	GetAllByUserID(ctx context.Context, userID domain.UserID) (*LabelGetAllOutput, error)
	Create(ctx context.Context, userID domain.UserID, input LabelCreateInput) (*Label, error)
	Delete(ctx context.Context, userID domain.UserID, labelID domain.LabelID) error
}

type labelUsecase struct {
	lr domain.LabelRepository
}

func NewLabelUsecase(lr domain.LabelRepository) LabelUsecase {
	return &labelUsecase{ lr }
}

func (lu *labelUsecase) GetAllByUserID(ctx context.Context, userID domain.UserID) (*LabelGetAllOutput, error) {
	labels, err := lu.lr.GetAllByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var results []*Label
	for _, l := range labels {
		results = append(results, fromLabelEntity(l))
	}

	return &LabelGetAllOutput{
		labels: results,
	}, nil
}

func (lu *labelUsecase) Create(ctx context.Context, userID domain.UserID, input LabelCreateInput) (*Label, error) {
	entityInput := input.toEntityInput()
	label, err := lu.lr.CreateLabel(ctx, userID, entityInput)
	if err != nil {
		return nil, err
	}

	return fromLabelEntity(label), nil
}

func (lu *labelUsecase) Delete(ctx context.Context, userID domain.UserID, labelID domain.LabelID) error {
	return lu.lr.DeleteLabel(ctx, userID, labelID)
}