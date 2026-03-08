package repository

import (
	"context"

	"github.com/erpang/post-sync/internal/domain"
	"gorm.io/gorm"
)

type ChannelRepository struct {
	db *gorm.DB
}

func NewChannelRepository(db *gorm.DB) *ChannelRepository {
	return &ChannelRepository{db: db}
}

func (r *ChannelRepository) CreateAccount(ctx context.Context, account *domain.ChannelAccount) error {
	return r.db.WithContext(ctx).Create(account).Error
}

func (r *ChannelRepository) UpdateAccount(ctx context.Context, account *domain.ChannelAccount) error {
	return r.db.WithContext(ctx).Save(account).Error
}

func (r *ChannelRepository) ListAccounts(ctx context.Context) ([]domain.ChannelAccount, error) {
	var accounts []domain.ChannelAccount
	err := r.db.WithContext(ctx).Order("created_at desc").Find(&accounts).Error
	return accounts, err
}

func (r *ChannelRepository) GetAccountByID(ctx context.Context, id string) (*domain.ChannelAccount, error) {
	var account domain.ChannelAccount
	err := r.db.WithContext(ctx).First(&account, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *ChannelRepository) CreateTarget(ctx context.Context, target *domain.ChannelTarget) error {
	return r.db.WithContext(ctx).Create(target).Error
}

func (r *ChannelRepository) UpdateTarget(ctx context.Context, target *domain.ChannelTarget) error {
	return r.db.WithContext(ctx).Save(target).Error
}

func (r *ChannelRepository) ListTargets(ctx context.Context) ([]domain.ChannelTarget, error) {
	var targets []domain.ChannelTarget
	err := r.db.WithContext(ctx).Order("created_at desc").Find(&targets).Error
	return targets, err
}

func (r *ChannelRepository) GetTargetByID(ctx context.Context, id string) (*domain.ChannelTarget, error) {
	var target domain.ChannelTarget
	err := r.db.WithContext(ctx).First(&target, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &target, nil
}

func (r *ChannelRepository) GetTargetsByIDs(ctx context.Context, ids []string) ([]domain.ChannelTarget, error) {
	var targets []domain.ChannelTarget
	err := r.db.WithContext(ctx).Find(&targets, "id IN ?", ids).Error
	return targets, err
}
