package repository

import (
	"context"

	"github.com/erpang/post-sync/internal/domain"
	"gorm.io/gorm"
)

type PublishRepository struct {
	db *gorm.DB
}

func NewPublishRepository(db *gorm.DB) *PublishRepository {
	return &PublishRepository{db: db}
}

func (r *PublishRepository) CreateJob(ctx context.Context, job *domain.PublishJob) error {
	return r.db.WithContext(ctx).Create(job).Error
}

func (r *PublishRepository) UpdateJob(ctx context.Context, job *domain.PublishJob) error {
	return r.db.WithContext(ctx).Save(job).Error
}

func (r *PublishRepository) ListJobs(ctx context.Context) ([]domain.PublishJob, error) {
	var jobs []domain.PublishJob
	err := r.db.WithContext(ctx).Order("created_at desc").Find(&jobs).Error
	return jobs, err
}

func (r *PublishRepository) GetJobByID(ctx context.Context, id string) (*domain.PublishJob, error) {
	var job domain.PublishJob
	err := r.db.WithContext(ctx).First(&job, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *PublishRepository) CreateDelivery(ctx context.Context, delivery *domain.DeliveryTask) error {
	return r.db.WithContext(ctx).Create(delivery).Error
}

func (r *PublishRepository) UpdateDelivery(ctx context.Context, delivery *domain.DeliveryTask) error {
	return r.db.WithContext(ctx).Save(delivery).Error
}

func (r *PublishRepository) GetDeliveryByID(ctx context.Context, id string) (*domain.DeliveryTask, error) {
	var delivery domain.DeliveryTask
	err := r.db.WithContext(ctx).First(&delivery, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &delivery, nil
}

func (r *PublishRepository) ListDeliveriesByJobID(ctx context.Context, jobID string) ([]domain.DeliveryTask, error) {
	var deliveries []domain.DeliveryTask
	err := r.db.WithContext(ctx).
		Order("created_at asc").
		Find(&deliveries, "publish_job_id = ?", jobID).Error
	return deliveries, err
}

func (r *PublishRepository) ExistsSuccessfulDuplicate(ctx context.Context, channelType, targetKey, bodyHash string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.DeliveryTask{}).
		Where("channel_type = ? AND target_key = ? AND body_hash = ? AND status = ?", channelType, targetKey, bodyHash, domain.DeliveryStatusSuccess).
		Count(&count).Error
	return count > 0, err
}

func (r *PublishRepository) HasJobsForContent(ctx context.Context, contentID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.PublishJob{}).
		Where("content_id = ?", contentID).
		Count(&count).Error
	return count > 0, err
}
