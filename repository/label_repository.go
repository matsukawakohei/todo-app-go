package repository

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/matsukawakohei/todo-app-go/domain"
	"gorm.io/gorm"
)

type labelModel struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"uniqueIndex:idx_labels_user_id"`
	UserID    uint      `gorm:"uniqueIndex:idx_labels_user_id"`
	User      userModel `gorm:"constraint:OnDelete:CASCADE;"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (lm *labelModel) toEntity() *domain.Label {
	return &domain.Label{
		ID:        domain.LabelID(lm.ID),
		Name:      lm.Name,
		CreatedAt: lm.CreatedAt,
		UpdatedAt: lm.UpdatedAt,
	}
}

func (ml labelModel) TableName() string {
	return "labels"
}

type labelRepository struct {
	db *gorm.DB
}

func NewLabelRepository(db *gorm.DB) domain.LabelRepository {
	return &labelRepository{db}
}

func (lr *labelRepository) GetAllByUserID(ctx context.Context, userID domain.UserID) ([]domain.Label, error) {
	var labels []domain.Label
	if result := lr.db.WithContext(ctx).Where("user_id = ?", userID).Order("id desc").Find(&labels); result.Error != nil {
		slog.ErrorContext(ctx, "GetAllByUserID: unexpected error", "error", result.Error)
		return []domain.Label{}, domain.ErrUnexpected
	}

	return labels, nil
}

func (lr *labelRepository) CreateLabel(ctx context.Context, userID domain.UserID, label domain.CreateLabel) (*domain.Label, error) {
	labelModel := labelModel{
		Name:   label.Name,
		UserID: uint(userID),
	}

	if err := lr.db.WithContext(ctx).Create(&labelModel).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, domain.ErrDuplicatedKey
		}
		slog.ErrorContext(ctx, "CreateLabel: unexpected error", "error", err)
		return nil, domain.ErrUnexpected
	}

	return labelModel.toEntity(), nil
}

func (lr *labelRepository) DeleteLabel(ctx context.Context, userID domain.UserID, labelID domain.LabelID) error {
	result := lr.db.WithContext(ctx).Where("user_id = ? and id = ?", userID, labelID).Delete(&labelModel{})
	if result.Error != nil {
		slog.ErrorContext(ctx, "DeleteLabel: unexpected error", "error", result.Error)
		return domain.ErrUnexpected
	}

	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}
