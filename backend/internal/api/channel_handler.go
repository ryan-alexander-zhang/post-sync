package api

import (
	"errors"
	"net/http"

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
	c.JSON(http.StatusOK, gin.H{"items": accounts})
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

	c.JSON(http.StatusCreated, account)
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

	c.JSON(http.StatusOK, account)
}

func (h *ChannelHandler) ListTargets(c *gin.Context) {
	targets, err := h.channelService.ListTargets(c.Request.Context())
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": targets})
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

	c.JSON(http.StatusCreated, target)
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

	c.JSON(http.StatusOK, target)
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
