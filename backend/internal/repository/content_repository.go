package repository

import (
	"context"

	"github.com/erpang/post-sync/internal/domain"
	"gorm.io/gorm"
)

type ContentRepository struct {
	db *gorm.DB
}

func NewContentRepository(db *gorm.DB) *ContentRepository {
	return &ContentRepository{db: db}
}

func (r *ContentRepository) Create(ctx context.Context, content *domain.Content) error {
	return r.db.WithContext(ctx).Create(content).Error
}

func (r *ContentRepository) List(ctx context.Context) ([]domain.Content, error) {
	var contents []domain.Content
	err := r.db.WithContext(ctx).
		Order("created_at desc").
		Find(&contents).Error
	return contents, err
}

func (r *ContentRepository) GetByID(ctx context.Context, id string) (*domain.Content, error) {
	var content domain.Content
	err := r.db.WithContext(ctx).First(&content, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &content, nil
}
