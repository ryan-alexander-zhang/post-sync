package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/erpang/post-sync/internal/channel"
	"github.com/erpang/post-sync/internal/config"
	"github.com/erpang/post-sync/internal/domain"
	"github.com/erpang/post-sync/internal/render"
	"github.com/erpang/post-sync/internal/repository"
	"github.com/erpang/post-sync/internal/util"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

const defaultTemplate = `{{ if .Content.Title }}# {{ .Content.Title }}{{ end }}

{{ .Content.BodyMarkdown }}

{{ with .Meta.tags }}Tags: {{ join . ", " }}{{ end }}`

type PublishService struct {
	contentRepository *repository.ContentRepository
	channelRepository *repository.ChannelRepository
	publishRepository *repository.PublishRepository
	registry          *channel.Registry
	renderer          *render.TemplateRenderer
	config            config.PublishConfig
}

type CreatePublishJobInput struct {
	ContentID    string   `json:"contentId"`
	TargetIDs    []string `json:"targetIds"`
	TemplateName string   `json:"templateName"`
}

func NewPublishService(
	contentRepository *repository.ContentRepository,
	channelRepository *repository.ChannelRepository,
	publishRepository *repository.PublishRepository,
	registry *channel.Registry,
	renderer *render.TemplateRenderer,
	cfg config.PublishConfig,
) *PublishService {
	return &PublishService{
		contentRepository: contentRepository,
		channelRepository: channelRepository,
		publishRepository: publishRepository,
		registry:          registry,
		renderer:          renderer,
		config:            cfg,
	}
}

func (s *PublishService) CreateJob(ctx context.Context, input CreatePublishJobInput) (*domain.PublishJob, error) {
	if strings.TrimSpace(input.ContentID) == "" || len(input.TargetIDs) == 0 {
		return nil, fmt.Errorf("%w: contentId and targetIds are required", ErrValidation)
	}

	content, err := s.contentRepository.GetByID(ctx, input.ContentID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	targets, err := s.channelRepository.GetTargetsByIDs(ctx, input.TargetIDs)
	if err != nil {
		return nil, err
	}
	if len(targets) != len(input.TargetIDs) {
		return nil, ErrNotFound
	}

	requestID := util.NewID()
	now := time.Now().UTC()
	job := &domain.PublishJob{
		ID:              util.NewID(),
		ContentID:       content.ID,
		RequestID:       requestID,
		TriggerSource:   "manual",
		Status:          domain.PublishStatusPending,
		TotalDeliveries: len(targets),
		CreatedAt:       now,
	}

	if err := s.publishRepository.CreateJob(ctx, job); err != nil {
		return nil, err
	}

	templateName := input.TemplateName
	if strings.TrimSpace(templateName) == "" {
		templateName = "default-telegram"
	}

	for _, target := range targets {
		account, err := s.channelRepository.GetAccountByID(ctx, target.ChannelAccountID)
		if err != nil {
			return nil, err
		}

		delivery := &domain.DeliveryTask{
			ID:               util.NewID(),
			PublishJobID:     job.ID,
			ContentID:        content.ID,
			ChannelAccountID: account.ID,
			ChannelTargetID:  target.ID,
			ChannelType:      account.ChannelType,
			TargetKey:        target.TargetKey,
			Status:           domain.DeliveryStatusPending,
			AttemptCount:     0,
			IdempotencyKey:   computeIdempotencyKey(account.ChannelType, target.TargetKey, content.BodyHash),
			BodyHash:         content.BodyHash,
			TemplateName:     templateName,
			RenderMode:       domain.RenderModeTelegram,
		}

		if err := s.publishRepository.CreateDelivery(ctx, delivery); err != nil {
			return nil, err
		}
	}

	go s.executeJob(job.ID)
	return job, nil
}

func (s *PublishService) ListJobs(ctx context.Context) ([]domain.PublishJob, error) {
	return s.publishRepository.ListJobs(ctx)
}

func (s *PublishService) GetJobDetail(ctx context.Context, id string) (*domain.PublishJob, []domain.DeliveryTask, error) {
	job, err := s.publishRepository.GetJobByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, ErrNotFound
	}
	if err != nil {
		return nil, nil, err
	}

	deliveries, err := s.publishRepository.ListDeliveriesByJobID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	return job, deliveries, nil
}

func (s *PublishService) RetryDelivery(ctx context.Context, id string) (*domain.DeliveryTask, error) {
	delivery, err := s.publishRepository.GetDeliveryByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if delivery.Status != domain.DeliveryStatusFailed {
		return nil, fmt.Errorf("%w: only failed deliveries can be retried", ErrValidation)
	}

	delivery.Status = domain.DeliveryStatusPending
	delivery.ErrorCode = ""
	delivery.ErrorMessage = ""
	delivery.ProviderResponseJSON = ""
	if err := s.publishRepository.UpdateDelivery(ctx, delivery); err != nil {
		return nil, err
	}

	go s.executeDelivery(delivery.ID)
	return delivery, nil
}

func (s *PublishService) executeJob(jobID string) {
	ctx := context.Background()
	job, err := s.publishRepository.GetJobByID(ctx, jobID)
	if err != nil {
		return
	}

	now := time.Now().UTC()
	job.Status = domain.PublishStatusProcessing
	job.StartedAt = &now
	_ = s.publishRepository.UpdateJob(ctx, job)

	deliveries, err := s.publishRepository.ListDeliveriesByJobID(ctx, jobID)
	if err != nil {
		job.Status = domain.PublishStatusFailed
		finished := time.Now().UTC()
		job.FinishedAt = &finished
		_ = s.publishRepository.UpdateJob(ctx, job)
		return
	}

	var (
		group, groupCtx = errgroup.WithContext(ctx)
		sem             = make(chan struct{}, max(1, s.config.MaxParallelism))
	)

	for _, delivery := range deliveries {
		deliveryID := delivery.ID
		group.Go(func() error {
			select {
			case sem <- struct{}{}:
			case <-groupCtx.Done():
				return groupCtx.Err()
			}
			defer func() { <-sem }()

			s.executeDelivery(deliveryID)
			return nil
		})
	}

	_ = group.Wait()
	s.refreshJobStatus(ctx, jobID)
}

func (s *PublishService) executeDelivery(deliveryID string) {
	ctx, cancel := context.WithTimeout(context.Background(), s.config.Timeout)
	defer cancel()

	delivery, err := s.publishRepository.GetDeliveryByID(ctx, deliveryID)
	if err != nil {
		return
	}

	startedAt := time.Now().UTC()
	delivery.Status = domain.DeliveryStatusProcessing
	delivery.StartedAt = &startedAt
	delivery.AttemptCount++
	_ = s.publishRepository.UpdateDelivery(ctx, delivery)

	content, err := s.contentRepository.GetByID(ctx, delivery.ContentID)
	if err != nil {
		s.failDelivery(ctx, delivery, "CONTENT_NOT_FOUND", err.Error())
		s.refreshJobStatus(context.Background(), delivery.PublishJobID)
		return
	}

	duplicate, err := s.publishRepository.ExistsSuccessfulDuplicate(ctx, delivery.ChannelType, delivery.TargetKey, delivery.BodyHash)
	if err != nil {
		s.failDelivery(ctx, delivery, "DEDUP_CHECK_FAILED", err.Error())
		s.refreshJobStatus(context.Background(), delivery.PublishJobID)
		return
	}
	if duplicate {
		finished := time.Now().UTC()
		delivery.Status = domain.DeliveryStatusSkippedDuplicate
		delivery.FinishedAt = &finished
		_ = s.publishRepository.UpdateDelivery(ctx, delivery)
		s.refreshJobStatus(context.Background(), delivery.PublishJobID)
		return
	}

	account, err := s.channelRepository.GetAccountByID(ctx, delivery.ChannelAccountID)
	if err != nil {
		s.failDelivery(ctx, delivery, "CHANNEL_ACCOUNT_NOT_FOUND", err.Error())
		s.refreshJobStatus(context.Background(), delivery.PublishJobID)
		return
	}

	target, err := s.channelRepository.GetTargetByID(ctx, delivery.ChannelTargetID)
	if err != nil {
		s.failDelivery(ctx, delivery, "CHANNEL_TARGET_NOT_FOUND", err.Error())
		s.refreshJobStatus(context.Background(), delivery.PublishJobID)
		return
	}

	driver, err := s.registry.MustGet(delivery.ChannelType)
	if err != nil {
		s.failDelivery(ctx, delivery, "CHANNEL_DRIVER_NOT_FOUND", err.Error())
		s.refreshJobStatus(context.Background(), delivery.PublishJobID)
		return
	}

	frontmatter := map[string]any{}
	if content.FrontmatterJSON != "" {
		_ = json.Unmarshal([]byte(content.FrontmatterJSON), &frontmatter)
	}

	renderContext := render.Context{Meta: frontmatter}
	renderContext.Content.ID = content.ID
	renderContext.Content.Title = content.Title
	renderContext.Content.BodyMarkdown = content.BodyMarkdown
	renderContext.Content.BodyPlain = content.BodyPlain
	renderContext.Channel.TargetName = target.TargetName
	renderContext.Channel.TargetKey = target.TargetKey

	renderedHTML, err := s.renderer.RenderMarkdown(delivery.TemplateName, defaultTemplate, renderContext)
	if err != nil {
		s.failDelivery(ctx, delivery, "RENDER_ERROR", err.Error())
		s.refreshJobStatus(context.Background(), delivery.PublishJobID)
		return
	}

	driverRendered, err := driver.Render(channel.RenderInput{
		TemplateName: delivery.TemplateName,
		ContentTitle: content.Title,
		ContentBody:  renderedHTML,
		Frontmatter:  frontmatter,
		TargetName:   target.TargetName,
		TargetKey:    target.TargetKey,
	})
	if err != nil {
		s.failDelivery(ctx, delivery, "RENDER_ERROR", err.Error())
		s.refreshJobStatus(context.Background(), delivery.PublishJobID)
		return
	}

	delivery.RenderedTitle = driverRendered.Title
	delivery.RenderedBody = driverRendered.Body
	delivery.RenderMode = driverRendered.RenderMode
	_ = s.publishRepository.UpdateDelivery(ctx, delivery)

	targetConfig := map[string]any{}
	if target.ConfigJSON != "" {
		_ = json.Unmarshal([]byte(target.ConfigJSON), &targetConfig)
	}

	sendResult, sendErr := s.sendWithRetry(ctx, driver, channel.SendRequest{
		SecretValue: os.Getenv(account.SecretRef),
		TargetKey:   target.TargetKey,
		Body:        driverRendered.Body,
		Config:      targetConfig,
	})
	if sendErr != nil {
		s.failDelivery(ctx, delivery, "CHANNEL_SEND_FAILED", sendErr.Error())
		s.refreshJobStatus(context.Background(), delivery.PublishJobID)
		return
	}

	finished := time.Now().UTC()
	delivery.Status = domain.DeliveryStatusSuccess
	delivery.ExternalMessageID = sendResult.ExternalMessageID
	delivery.ProviderResponseJSON = sendResult.ProviderResponse
	delivery.FinishedAt = &finished
	delivery.ErrorCode = ""
	delivery.ErrorMessage = ""
	_ = s.publishRepository.UpdateDelivery(ctx, delivery)
	s.refreshJobStatus(context.Background(), delivery.PublishJobID)
}

func (s *PublishService) sendWithRetry(ctx context.Context, driver channel.Driver, request channel.SendRequest) (channel.SendResult, error) {
	var result channel.SendResult
	var err error
	backoffs := []time.Duration{0, time.Second, 3 * time.Second}

	for attempt, backoff := range backoffs {
		if backoff > 0 {
			time.Sleep(backoff)
		}

		result, err = driver.Send(ctx, request)
		if err == nil {
			return result, nil
		}

		if attempt == len(backoffs)-1 || !isRetryableError(err) {
			return channel.SendResult{}, err
		}
	}

	return channel.SendResult{}, err
}

func (s *PublishService) failDelivery(ctx context.Context, delivery *domain.DeliveryTask, code, message string) {
	finished := time.Now().UTC()
	delivery.Status = domain.DeliveryStatusFailed
	delivery.ErrorCode = code
	delivery.ErrorMessage = message
	delivery.FinishedAt = &finished
	_ = s.publishRepository.UpdateDelivery(ctx, delivery)
}

func (s *PublishService) refreshJobStatus(ctx context.Context, jobID string) {
	job, err := s.publishRepository.GetJobByID(ctx, jobID)
	if err != nil {
		return
	}
	deliveries, err := s.publishRepository.ListDeliveriesByJobID(ctx, jobID)
	if err != nil {
		return
	}

	var successCount, failedCount, skippedCount int
	for _, delivery := range deliveries {
		switch delivery.Status {
		case domain.DeliveryStatusSuccess:
			successCount++
		case domain.DeliveryStatusFailed:
			failedCount++
		case domain.DeliveryStatusSkippedDuplicate:
			skippedCount++
		}
	}

	job.SuccessCount = successCount
	job.FailedCount = failedCount
	job.SkippedCount = skippedCount
	job.TotalDeliveries = len(deliveries)
	job.Status = aggregateJobStatus(deliveries)
	if job.Status == domain.PublishStatusSuccess || job.Status == domain.PublishStatusFailed || job.Status == domain.PublishStatusPartialSuccess {
		finished := time.Now().UTC()
		job.FinishedAt = &finished
	}

	_ = s.publishRepository.UpdateJob(ctx, job)
}

func aggregateJobStatus(deliveries []domain.DeliveryTask) string {
	if len(deliveries) == 0 {
		return domain.PublishStatusFailed
	}

	var (
		pendingOrProcessing bool
		successCount        int
		failedCount         int
		skippedCount        int
	)

	for _, delivery := range deliveries {
		switch delivery.Status {
		case domain.DeliveryStatusPending, domain.DeliveryStatusProcessing:
			pendingOrProcessing = true
		case domain.DeliveryStatusSuccess:
			successCount++
		case domain.DeliveryStatusFailed:
			failedCount++
		case domain.DeliveryStatusSkippedDuplicate:
			skippedCount++
		}
	}

	if pendingOrProcessing {
		return domain.PublishStatusProcessing
	}
	if failedCount == len(deliveries) {
		return domain.PublishStatusFailed
	}
	if failedCount > 0 && successCount+skippedCount > 0 {
		return domain.PublishStatusPartialSuccess
	}
	return domain.PublishStatusSuccess
}

func computeIdempotencyKey(channelType, targetKey, bodyHash string) string {
	sum := sha256.Sum256([]byte(channelType + ":" + targetKey + ":" + bodyHash))
	return hex.EncodeToString(sum[:])
}

func isRetryableError(err error) bool {
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "timeout") ||
		strings.Contains(message, "tempor") ||
		strings.Contains(message, "rate limit") ||
		strings.Contains(message, "5")
}

func max(left, right int) int {
	if left > right {
		return left
	}
	return right
}
