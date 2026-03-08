package api

import (
	"errors"
	"net/http"

	"github.com/erpang/post-sync/internal/service"
	"github.com/gin-gonic/gin"
)

type PublishHandler struct {
	publishService *service.PublishService
}

func NewPublishHandler(publishService *service.PublishService) *PublishHandler {
	return &PublishHandler{publishService: publishService}
}

func (h *PublishHandler) RegisterRoutes(group *gin.RouterGroup) {
	group.POST("/publish-jobs", h.CreateJob)
	group.GET("/publish-jobs", h.ListJobs)
	group.GET("/publish-jobs/:id", h.GetJobByID)
	group.POST("/delivery-tasks/:id/retry", h.RetryDelivery)
}

func (h *PublishHandler) CreateJob(c *gin.Context) {
	var request service.CreatePublishJobInput
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}

	job, err := h.publishService.CreateJob(c.Request.Context(), request)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			writeError(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "content or target not found")
		default:
			writeServiceError(c, err)
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"jobId":  job.ID,
		"status": job.Status,
	})
}

func (h *PublishHandler) ListJobs(c *gin.Context) {
	jobs, err := h.publishService.ListJobs(c.Request.Context())
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": jobs})
}

func (h *PublishHandler) GetJobByID(c *gin.Context) {
	job, deliveries, err := h.publishService.GetJobDetail(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeError(c, http.StatusNotFound, "PUBLISH_JOB_NOT_FOUND", "publish job not found")
			return
		}
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"job":        job,
		"deliveries": deliveries,
	})
}

func (h *PublishHandler) RetryDelivery(c *gin.Context) {
	delivery, err := h.publishService.RetryDelivery(c.Request.Context(), c.Param("id"))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			writeError(c, http.StatusNotFound, "DELIVERY_TASK_NOT_FOUND", "delivery task not found")
		default:
			writeServiceError(c, err)
		}
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"deliveryId": delivery.ID,
		"status":     delivery.Status,
	})
}
