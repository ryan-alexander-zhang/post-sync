package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/erpang/post-sync/internal/domain"
	"github.com/erpang/post-sync/internal/parser"
	"github.com/erpang/post-sync/internal/repository"
	"github.com/erpang/post-sync/internal/util"
	"gorm.io/gorm"
)

type ContentService struct {
	contentRepository *repository.ContentRepository
	publishRepository *repository.PublishRepository
}

func NewContentService(
	contentRepository *repository.ContentRepository,
	publishRepository *repository.PublishRepository,
) *ContentService {
	return &ContentService{
		contentRepository: contentRepository,
		publishRepository: publishRepository,
	}
}

func (s *ContentService) Upload(ctx context.Context, filename string, raw []byte) (*domain.Content, error) {
	if strings.TrimSpace(filename) == "" {
		return nil, fmt.Errorf("%w: filename is required", ErrValidation)
	}
	if len(raw) == 0 {
		return nil, fmt.Errorf("%w: file is empty", ErrValidation)
	}

	parsed, err := parser.ParseMarkdown(raw)
	if err != nil {
		return nil, fmt.Errorf("%w: parse markdown: %v", ErrValidation, err)
	}

	existing, err := s.contentRepository.GetByBodyHash(ctx, parsed.BodyHash)
	if err == nil {
		return nil, fmt.Errorf(
			"%w: duplicate content upload; existing content id=%s",
			ErrValidation,
			existing.ID,
		)
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	content := &domain.Content{
		ID:               util.NewID(),
		SourceFilename:   filename,
		OriginalMarkdown: string(raw),
		FrontmatterJSON:  parsed.FrontmatterJSON,
		Title:            parsed.Title,
		BodyMarkdown:     parsed.BodyMarkdown,
		BodyPlain:        parsed.BodyPlain,
		BodyHash:         parsed.BodyHash,
	}

	if err := s.contentRepository.Create(ctx, content); err != nil {
		return nil, err
	}

	return content, nil
}

func (s *ContentService) List(ctx context.Context) ([]domain.Content, error) {
	return s.contentRepository.List(ctx)
}

func (s *ContentService) GetByID(ctx context.Context, id string) (*domain.Content, error) {
	content, err := s.contentRepository.GetByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return content, nil
}

func (s *ContentService) DeleteByID(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("%w: content id is required", ErrValidation)
	}

	_, err := s.contentRepository.GetByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	hasJobs, err := s.publishRepository.HasJobsForContent(ctx, id)
	if err != nil {
		return err
	}
	if hasJobs {
		return fmt.Errorf("%w: content with publish history cannot be deleted", ErrValidation)
	}

	return s.contentRepository.DeleteByID(ctx, id)
}
