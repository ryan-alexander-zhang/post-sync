package api

import (
	"errors"
	"net/http"

	"github.com/erpang/post-sync/internal/service"
	"github.com/gin-gonic/gin"
)

func writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrValidation):
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	case errors.Is(err, service.ErrConfiguration):
		writeError(c, http.StatusBadRequest, "CONFIG_ERROR", err.Error())
	default:
		writeError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}

func writeError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{
		"code":    code,
		"message": message,
	})
}
