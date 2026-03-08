package api

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func NewRouter(database *gorm.DB) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	router.GET("/api/v1/system/info", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"database": database.Name(),
			"status":   "ready",
		})
	})

	return router
}
