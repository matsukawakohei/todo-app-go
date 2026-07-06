package repository

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/matsukawakohei/todo-app-go/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const MaxLabelCount = 100

type labelModel struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"uniqueIndex:idx_labels_user_id"`
	UserID    uint   `gorm:"uniqueIndex:idx_labels_user_id"`
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

func (lm labelModel) TableName() string {
	return "labels"
}

type labelRepository struct {
	db *gorm.DB
}

func NewLabelRepository(db *gorm.DB) domain.LabelRepository {
	return &labelRepository{db}
}

func (lr *labelRepository) GetAllByUserID(ctx context.Context, userID domain.UserID) ([]*domain.Label, error) {
	var models []labelModel
	if result := lr.db.WithContext(ctx).Where("user_id = ?", userID).Order("id desc").Find(&models); result.Error != nil {
		slog.ErrorContext(ctx, "GetAllByUserID: unexpected error", "error", result.Error)
		return []*domain.Label{}, domain.ErrUnexpected
	}

	var labels []*domain.Label
	for _, l := range models {
		labels = append(labels, l.toEntity())
	}

	return labels, nil
}

func (lr *labelRepository) CreateLabel(ctx context.Context, userID domain.UserID, label domain.CreateLabel) (*domain.Label, error) {
	model := labelModel{
		Name:   label.Name,
		UserID: uint(userID),
	}

	err := lr.db.Transaction(func(tx *gorm.DB) error {
		tx = tx.WithContext(ctx)
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&userModel{}, userID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrNotFound
			}
			slog.ErrorContext(ctx, "CreateLabel: failed lock user", "error", err, "userID", userID)
			return domain.ErrUnexpected
		}
		var labelCount int64
		if err := tx.Model(&labelModel{}).Where("user_id = ?", userID).Count(&labelCount).Error; err != nil {
			slog.ErrorContext(ctx, "CreateLabel: failed count label", "error", err)
			return domain.ErrUnexpected
		}

		if labelCount >= MaxLabelCount {
			slog.WarnContext(ctx, "label count is limit", "userID", userID)
			return domain.ErrLimitExceeded
		}

		if err := tx.Create(&model).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return domain.ErrDuplicatedKey
			}
			slog.ErrorContext(ctx, "CreateLabel: unexpected error", "error", err)
			return domain.ErrUnexpected
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return model.toEntity(), nil
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
