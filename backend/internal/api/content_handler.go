package api

import (
	"errors"
	"io"
	"net/http"

	"github.com/erpang/post-sync/internal/service"
	"github.com/gin-gonic/gin"
)

type ContentHandler struct {
	contentService *service.ContentService
}

func NewContentHandler(contentService *service.ContentService) *ContentHandler {
	return &ContentHandler{contentService: contentService}
}

func (h *ContentHandler) RegisterRoutes(group *gin.RouterGroup) {
	group.POST("/contents/upload", h.Upload)
	group.GET("/contents", h.List)
	group.GET("/contents/:id", h.GetByID)
	group.DELETE("/contents/:id", h.DeleteByID)
}

func (h *ContentHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_FILE", "file is required")
		return
	}

	opened, err := file.Open()
	if err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_FILE", "cannot open uploaded file")
		return
	}
	defer opened.Close()

	data, err := io.ReadAll(opened)
	if err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_FILE", "cannot read uploaded file")
		return
	}

	content, err := h.contentService.Upload(c.Request.Context(), file.Filename, data)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, content)
}

func (h *ContentHandler) List(c *gin.Context) {
	contents, err := h.contentService.List(c.Request.Context())
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": contents})
}

func (h *ContentHandler) GetByID(c *gin.Context) {
	content, err := h.contentService.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeError(c, http.StatusNotFound, "CONTENT_NOT_FOUND", "content not found")
			return
		}
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, content)
}

func (h *ContentHandler) DeleteByID(c *gin.Context) {
	err := h.contentService.DeleteByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeError(c, http.StatusNotFound, "CONTENT_NOT_FOUND", "content not found")
			return
		}
		writeServiceError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
