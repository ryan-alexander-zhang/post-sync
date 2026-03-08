package domain

import "time"

type Content struct {
	ID               string    `gorm:"primaryKey;size:32" json:"id"`
	SourceFilename   string    `gorm:"size:255;not null" json:"sourceFilename"`
	OriginalMarkdown string    `gorm:"type:text;not null" json:"originalMarkdown"`
	FrontmatterJSON  string    `gorm:"type:text" json:"frontmatterJson"`
	Title            string    `gorm:"size:255" json:"title"`
	BodyMarkdown     string    `gorm:"type:text;not null;index:idx_contents_body_hash" json:"bodyMarkdown"`
	BodyPlain        string    `gorm:"type:text" json:"bodyPlain"`
	BodyHash         string    `gorm:"size:64;not null;index:idx_contents_body_hash" json:"bodyHash"`
	CreatedAt        time.Time `json:"createdAt"`
}

type ChannelAccount struct {
	ID          string    `gorm:"primaryKey;size:32" json:"id"`
	ChannelType string    `gorm:"size:50;not null;index:idx_channel_accounts_type" json:"channelType"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Enabled     bool      `gorm:"not null;default:true" json:"enabled"`
	SecretRef   string    `gorm:"size:100;not null" json:"secretRef"`
	ConfigJSON  string    `gorm:"type:text" json:"configJson"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type ChannelTarget struct {
	ID               string    `gorm:"primaryKey;size:32" json:"id"`
	ChannelAccountID string    `gorm:"size:32;not null;index:idx_channel_targets_account_id" json:"channelAccountId"`
	TargetType       string    `gorm:"size:50;not null" json:"targetType"`
	TargetKey        string    `gorm:"size:255;not null;index:idx_channel_targets_target_key" json:"targetKey"`
	TargetName       string    `gorm:"size:100;not null" json:"targetName"`
	Enabled          bool      `gorm:"not null;default:true" json:"enabled"`
	ConfigJSON       string    `gorm:"type:text" json:"configJson"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type PublishJob struct {
	ID              string     `gorm:"primaryKey;size:32" json:"id"`
	ContentID       string     `gorm:"size:32;not null;index:idx_publish_jobs_content_id" json:"contentId"`
	RequestID       string     `gorm:"size:100;not null" json:"requestId"`
	TriggerSource   string     `gorm:"size:50;not null" json:"triggerSource"`
	Status          string     `gorm:"size:30;not null;index:idx_publish_jobs_status" json:"status"`
	TotalDeliveries int        `gorm:"not null" json:"totalDeliveries"`
	SuccessCount    int        `gorm:"not null" json:"successCount"`
	FailedCount     int        `gorm:"not null" json:"failedCount"`
	SkippedCount    int        `gorm:"not null" json:"skippedCount"`
	CreatedAt       time.Time  `gorm:"index:idx_publish_jobs_created_at" json:"createdAt"`
	StartedAt       *time.Time `json:"startedAt"`
	FinishedAt      *time.Time `json:"finishedAt"`
}

type DeliveryTask struct {
	ID                   string     `gorm:"primaryKey;size:32" json:"id"`
	PublishJobID         string     `gorm:"size:32;not null;index:idx_delivery_tasks_publish_job_id" json:"publishJobId"`
	ContentID            string     `gorm:"size:32;not null" json:"contentId"`
	ChannelAccountID     string     `gorm:"size:32;not null" json:"channelAccountId"`
	ChannelTargetID      string     `gorm:"size:32;not null" json:"channelTargetId"`
	ChannelType          string     `gorm:"size:50;not null;index:idx_delivery_tasks_lookup_dedup" json:"channelType"`
	TargetKey            string     `gorm:"size:255;not null;index:idx_delivery_tasks_lookup_dedup" json:"targetKey"`
	Status               string     `gorm:"size:30;not null;index:idx_delivery_tasks_status;index:idx_delivery_tasks_lookup_dedup" json:"status"`
	AttemptCount         int        `gorm:"not null" json:"attemptCount"`
	IdempotencyKey       string     `gorm:"size:64;not null" json:"idempotencyKey"`
	BodyHash             string     `gorm:"size:64;not null;index:idx_delivery_tasks_lookup_dedup" json:"bodyHash"`
	TemplateName         string     `gorm:"size:100;not null" json:"templateName"`
	RenderMode           string     `gorm:"size:30;not null" json:"renderMode"`
	RenderedTitle        string     `gorm:"type:text" json:"renderedTitle"`
	RenderedBody         string     `gorm:"type:text" json:"renderedBody"`
	ExternalMessageID    string     `gorm:"size:255" json:"externalMessageId"`
	ErrorCode            string     `gorm:"size:100" json:"errorCode"`
	ErrorMessage         string     `gorm:"type:text" json:"errorMessage"`
	ProviderResponseJSON string     `gorm:"type:text" json:"providerResponseJson"`
	CreatedAt            time.Time  `gorm:"index:idx_delivery_tasks_created_at" json:"createdAt"`
	StartedAt            *time.Time `json:"startedAt"`
	FinishedAt           *time.Time `json:"finishedAt"`
}

const (
	PublishStatusPending        = "PENDING"
	PublishStatusProcessing     = "PROCESSING"
	PublishStatusSuccess        = "SUCCESS"
	PublishStatusPartialSuccess = "PARTIAL_SUCCESS"
	PublishStatusFailed         = "FAILED"

	DeliveryStatusPending          = "PENDING"
	DeliveryStatusProcessing       = "PROCESSING"
	DeliveryStatusSuccess          = "SUCCESS"
	DeliveryStatusFailed           = "FAILED"
	DeliveryStatusSkippedDuplicate = "SKIPPED_DUPLICATE"

	ChannelTypeTelegram     = "telegram"
	ChannelTypeFeishu       = "feishu"
	TargetTypeTelegramGrp   = "telegram_group"
	TargetTypeTelegramTopic = "telegram_topic"
	TargetTypeFeishuChat    = "feishu_chat"
	RenderModeTelegram      = "telegram_html"
	RenderModeFeishuText    = "feishu_text"
	RenderModeFeishuPost    = "feishu_post"
)
