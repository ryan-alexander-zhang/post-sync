package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/erpang/post-sync/internal/channel"
	"github.com/erpang/post-sync/internal/domain"
	"github.com/erpang/post-sync/internal/repository"
	"github.com/erpang/post-sync/internal/util"
	"gorm.io/gorm"
)

type ChannelService struct {
	channelRepository *repository.ChannelRepository
	registry          *channel.Registry
}

type CreateChannelAccountInput struct {
	ChannelType string         `json:"channelType"`
	Name        string         `json:"name"`
	Enabled     *bool          `json:"enabled"`
	SecretRef   string         `json:"secretRef"`
	Config      map[string]any `json:"config"`
}

type UpdateChannelAccountInput struct {
	Name      *string        `json:"name"`
	Enabled   *bool          `json:"enabled"`
	SecretRef *string        `json:"secretRef"`
	Config    map[string]any `json:"config"`
}

type CreateChannelTargetInput struct {
	ChannelAccountID string         `json:"channelAccountId"`
	TargetType       string         `json:"targetType"`
	TargetKey        string         `json:"targetKey"`
	TargetName       string         `json:"targetName"`
	Enabled          *bool          `json:"enabled"`
	Config           map[string]any `json:"config"`
}

type UpdateChannelTargetInput struct {
	TargetName *string        `json:"targetName"`
	Enabled    *bool          `json:"enabled"`
	Config     map[string]any `json:"config"`
}

func NewChannelService(channelRepository *repository.ChannelRepository, registry *channel.Registry) *ChannelService {
	return &ChannelService{
		channelRepository: channelRepository,
		registry:          registry,
	}
}

func (s *ChannelService) ListAccounts(ctx context.Context) ([]domain.ChannelAccount, error) {
	return s.channelRepository.ListAccounts(ctx)
}

func (s *ChannelService) CreateAccount(ctx context.Context, input CreateChannelAccountInput) (*domain.ChannelAccount, error) {
	if strings.TrimSpace(input.ChannelType) == "" || strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.SecretRef) == "" {
		return nil, fmt.Errorf("%w: channelType, name and secretRef are required", ErrValidation)
	}

	driver, err := s.registry.MustGet(input.ChannelType)
	if err != nil {
		return nil, fmt.Errorf("%w: unsupported channel type", ErrValidation)
	}

	if err := driver.ValidateAccount(input.Config, input.SecretRef); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrValidation, err)
	}

	configJSON, err := marshalConfig(input.Config)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid config", ErrValidation)
	}

	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}

	account := &domain.ChannelAccount{
		ID:          util.NewID(),
		ChannelType: input.ChannelType,
		Name:        strings.TrimSpace(input.Name),
		Enabled:     enabled,
		SecretRef:   strings.TrimSpace(input.SecretRef),
		ConfigJSON:  configJSON,
	}

	if err := s.channelRepository.CreateAccount(ctx, account); err != nil {
		return nil, err
	}
	return account, nil
}

func (s *ChannelService) UpdateAccount(ctx context.Context, id string, input UpdateChannelAccountInput) (*domain.ChannelAccount, error) {
	account, err := s.channelRepository.GetAccountByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		account.Name = strings.TrimSpace(*input.Name)
	}
	if input.Enabled != nil {
		account.Enabled = *input.Enabled
	}
	if input.SecretRef != nil {
		account.SecretRef = strings.TrimSpace(*input.SecretRef)
	}
	if input.Config != nil {
		account.ConfigJSON, err = marshalConfig(input.Config)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid config", ErrValidation)
		}
	}

	driver, err := s.registry.MustGet(account.ChannelType)
	if err != nil {
		return nil, fmt.Errorf("%w: unsupported channel type", ErrValidation)
	}
	if err := driver.ValidateAccount(input.Config, account.SecretRef); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrValidation, err)
	}

	if err := s.channelRepository.UpdateAccount(ctx, account); err != nil {
		return nil, err
	}

	return account, nil
}

func (s *ChannelService) ListTargets(ctx context.Context) ([]domain.ChannelTarget, error) {
	return s.channelRepository.ListTargets(ctx)
}

func (s *ChannelService) CreateTarget(ctx context.Context, input CreateChannelTargetInput) (*domain.ChannelTarget, error) {
	if strings.TrimSpace(input.ChannelAccountID) == "" || strings.TrimSpace(input.TargetKey) == "" || strings.TrimSpace(input.TargetName) == "" {
		return nil, fmt.Errorf("%w: channelAccountId, targetKey and targetName are required", ErrValidation)
	}

	account, err := s.channelRepository.GetAccountByID(ctx, input.ChannelAccountID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	driver, err := s.registry.MustGet(account.ChannelType)
	if err != nil {
		return nil, fmt.Errorf("%w: unsupported channel type", ErrValidation)
	}
	if err := driver.ValidateTarget(input.Config, input.TargetKey); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrValidation, err)
	}

	configJSON, err := marshalConfig(input.Config)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid config", ErrValidation)
	}

	targetType := input.TargetType
	if strings.TrimSpace(targetType) == "" {
		targetType = domain.TargetTypeTelegramGrp
	}

	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}

	target := &domain.ChannelTarget{
		ID:               util.NewID(),
		ChannelAccountID: account.ID,
		TargetType:       targetType,
		TargetKey:        strings.TrimSpace(input.TargetKey),
		TargetName:       strings.TrimSpace(input.TargetName),
		Enabled:          enabled,
		ConfigJSON:       configJSON,
	}

	if err := s.channelRepository.CreateTarget(ctx, target); err != nil {
		return nil, err
	}

	return target, nil
}

func (s *ChannelService) UpdateTarget(ctx context.Context, id string, input UpdateChannelTargetInput) (*domain.ChannelTarget, error) {
	target, err := s.channelRepository.GetTargetByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if input.TargetName != nil {
		target.TargetName = strings.TrimSpace(*input.TargetName)
	}
	if input.Enabled != nil {
		target.Enabled = *input.Enabled
	}
	if input.Config != nil {
		target.ConfigJSON, err = marshalConfig(input.Config)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid config", ErrValidation)
		}
	}

	account, err := s.channelRepository.GetAccountByID(ctx, target.ChannelAccountID)
	if err != nil {
		return nil, err
	}
	driver, err := s.registry.MustGet(account.ChannelType)
	if err != nil {
		return nil, fmt.Errorf("%w: unsupported channel type", ErrValidation)
	}

	configMap := map[string]any{}
	if input.Config != nil {
		configMap = input.Config
	}
	if err := driver.ValidateTarget(configMap, target.TargetKey); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrValidation, err)
	}

	if err := s.channelRepository.UpdateTarget(ctx, target); err != nil {
		return nil, err
	}

	return target, nil
}

func (s *ChannelService) DeleteAccount(ctx context.Context, id string) error {
	_, err := s.channelRepository.GetAccountByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	targetCount, err := s.channelRepository.CountTargetsByAccountID(ctx, id)
	if err != nil {
		return err
	}
	if targetCount > 0 {
		return fmt.Errorf("%w: delete targets before deleting the account", ErrValidation)
	}

	return s.channelRepository.DeleteAccount(ctx, id)
}

func (s *ChannelService) DeleteTarget(ctx context.Context, id string) error {
	_, err := s.channelRepository.GetTargetByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	return s.channelRepository.DeleteTarget(ctx, id)
}

func marshalConfig(config map[string]any) (string, error) {
	if len(config) == 0 {
		return "{}", nil
	}

	data, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
