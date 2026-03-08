package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/erpang/post-sync/internal/domain"
	"github.com/erpang/post-sync/internal/service"
	"github.com/gin-gonic/gin"
)

type ChannelHandler struct {
	channelService *service.ChannelService
}

func NewChannelHandler(channelService *service.ChannelService) *ChannelHandler {
	return &ChannelHandler{channelService: channelService}
}

func (h *ChannelHandler) RegisterRoutes(group *gin.RouterGroup) {
	group.GET("/channel-accounts", h.ListAccounts)
	group.POST("/channel-accounts", h.CreateAccount)
	group.PATCH("/channel-accounts/:id", h.UpdateAccount)
	group.DELETE("/channel-accounts/:id", h.DeleteAccount)
	group.GET("/channel-targets", h.ListTargets)
	group.POST("/channel-targets", h.CreateTarget)
	group.PATCH("/channel-targets/:id", h.UpdateTarget)
	group.DELETE("/channel-targets/:id", h.DeleteTarget)
}

func (h *ChannelHandler) ListAccounts(c *gin.Context) {
	accounts, err := h.channelService.ListAccounts(c.Request.Context())
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": sanitizeAccounts(accounts)})
}

func (h *ChannelHandler) CreateAccount(c *gin.Context) {
	var request service.CreateChannelAccountInput
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}

	account, err := h.channelService.CreateAccount(c.Request.Context(), request)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, sanitizeAccount(*account))
}

func (h *ChannelHandler) UpdateAccount(c *gin.Context) {
	var request service.UpdateChannelAccountInput
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}

	account, err := h.channelService.UpdateAccount(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeError(c, http.StatusNotFound, "CHANNEL_ACCOUNT_NOT_FOUND", "channel account not found")
			return
		}
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, sanitizeAccount(*account))
}

func (h *ChannelHandler) ListTargets(c *gin.Context) {
	targets, err := h.channelService.ListTargets(c.Request.Context())
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": sanitizeTargets(targets)})
}

func (h *ChannelHandler) CreateTarget(c *gin.Context) {
	var request service.CreateChannelTargetInput
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}

	target, err := h.channelService.CreateTarget(c.Request.Context(), request)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeError(c, http.StatusNotFound, "CHANNEL_ACCOUNT_NOT_FOUND", "channel account not found")
			return
		}
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, sanitizeTarget(*target))
}

func (h *ChannelHandler) UpdateTarget(c *gin.Context) {
	var request service.UpdateChannelTargetInput
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}

	target, err := h.channelService.UpdateTarget(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeError(c, http.StatusNotFound, "CHANNEL_TARGET_NOT_FOUND", "channel target not found")
			return
		}
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, sanitizeTarget(*target))
}

func (h *ChannelHandler) DeleteAccount(c *gin.Context) {
	err := h.channelService.DeleteAccount(c.Request.Context(), c.Param("id"))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			writeError(c, http.StatusNotFound, "CHANNEL_ACCOUNT_NOT_FOUND", "channel account not found")
		default:
			writeServiceError(c, err)
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ChannelHandler) DeleteTarget(c *gin.Context) {
	err := h.channelService.DeleteTarget(c.Request.Context(), c.Param("id"))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			writeError(c, http.StatusNotFound, "CHANNEL_TARGET_NOT_FOUND", "channel target not found")
		default:
			writeServiceError(c, err)
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func sanitizeAccounts(items []domain.ChannelAccount) []domain.ChannelAccount {
	sanitized := make([]domain.ChannelAccount, 0, len(items))
	for _, item := range items {
		sanitized = append(sanitized, sanitizeAccount(item))
	}
	return sanitized
}

func sanitizeTargets(items []domain.ChannelTarget) []domain.ChannelTarget {
	sanitized := make([]domain.ChannelTarget, 0, len(items))
	for _, item := range items {
		sanitized = append(sanitized, sanitizeTarget(item))
	}
	return sanitized
}

func sanitizeAccount(account domain.ChannelAccount) domain.ChannelAccount {
	if account.ChannelType != domain.ChannelTypePersonalFeishu {
		return account
	}

	config := parseJSONConfig(account.ConfigJSON)
	delete(config, "webhookUrl")
	delete(config, "signSecret")
	account.ConfigJSON = marshalSanitizedConfig(config)
	return account
}

func sanitizeTarget(target domain.ChannelTarget) domain.ChannelTarget {
	if target.TargetType != domain.TargetTypePersonalFeishuWebhook {
		return target
	}

	config := parseJSONConfig(target.ConfigJSON)
	delete(config, "webhookUrl")
	target.ConfigJSON = marshalSanitizedConfig(config)
	return target
}

func parseJSONConfig(raw string) map[string]any {
	if raw == "" {
		return map[string]any{}
	}
	config := map[string]any{}
	if err := json.Unmarshal([]byte(raw), &config); err != nil {
		return map[string]any{}
	}
	return config
}

func marshalSanitizedConfig(config map[string]any) string {
	if len(config) == 0 {
		return "{}"
	}
	data, err := json.Marshal(config)
	if err != nil {
		return "{}"
	}
	return string(data)
}
