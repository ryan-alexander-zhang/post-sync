package service

import (
	"context"
	"errors"
	"testing"

	"github.com/erpang/post-sync/internal/domain"
	"github.com/erpang/post-sync/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestContentServiceDeleteByIDDeletesUnpublishedContent(t *testing.T) {
	db := openTestDB(t)
	contentRepository := repository.NewContentRepository(db)
	publishRepository := repository.NewPublishRepository(db)
	service := NewContentService(contentRepository, publishRepository)

	content := &domain.Content{
		ID:               "content_1",
		SourceFilename:   "sample.md",
		OriginalMarkdown: "body",
		BodyMarkdown:     "body",
		BodyHash:         "hash_1",
	}
	if err := contentRepository.Create(context.Background(), content); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := service.DeleteByID(context.Background(), content.ID); err != nil {
		t.Fatalf("DeleteByID() error = %v", err)
	}

	_, err := contentRepository.GetByID(context.Background(), content.ID)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("GetByID() error = %v, want record not found", err)
	}
}

func TestContentServiceDeleteByIDRejectsPublishedContent(t *testing.T) {
	db := openTestDB(t)
	contentRepository := repository.NewContentRepository(db)
	publishRepository := repository.NewPublishRepository(db)
	service := NewContentService(contentRepository, publishRepository)

	content := &domain.Content{
		ID:               "content_2",
		SourceFilename:   "sample.md",
		OriginalMarkdown: "body",
		BodyMarkdown:     "body",
		BodyHash:         "hash_2",
	}
	if err := contentRepository.Create(context.Background(), content); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	job := &domain.PublishJob{
		ID:              "job_1",
		ContentID:       content.ID,
		RequestID:       "req_1",
		TriggerSource:   "manual",
		Status:          domain.PublishStatusSuccess,
		TotalDeliveries: 1,
	}
	if err := publishRepository.CreateJob(context.Background(), job); err != nil {
		t.Fatalf("CreateJob() error = %v", err)
	}

	err := service.DeleteByID(context.Background(), content.ID)
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("DeleteByID() error = %v, want validation error", err)
	}
}

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}

	if err := db.AutoMigrate(&domain.Content{}, &domain.PublishJob{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	return db
}
