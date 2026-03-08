package api

import (
	"github.com/erpang/post-sync/internal/channel"
	"github.com/erpang/post-sync/internal/channel/telegram"
	"github.com/erpang/post-sync/internal/config"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/erpang/post-sync/internal/repository"
	"github.com/erpang/post-sync/internal/render"
	"github.com/erpang/post-sync/internal/service"
)

func NewRouter(database *gorm.DB, cfg config.Config) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.Use(corsMiddleware())

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

	contentRepository := repository.NewContentRepository(database)
	channelRepository := repository.NewChannelRepository(database)
	publishRepository := repository.NewPublishRepository(database)
	driverRegistry := channel.NewRegistry(telegram.New())

	contentService := service.NewContentService(contentRepository)
	channelService := service.NewChannelService(channelRepository, driverRegistry)
	publishService := service.NewPublishService(
		contentRepository,
		channelRepository,
		publishRepository,
		driverRegistry,
		render.NewTemplateRenderer(),
		cfg.PublishConfig,
	)

	contentHandler := NewContentHandler(contentService)
	channelHandler := NewChannelHandler(channelService)
	publishHandler := NewPublishHandler(publishService)

	apiGroup := router.Group("/api/v1")
	contentHandler.RegisterRoutes(apiGroup)
	channelHandler.RegisterRoutes(apiGroup)
	publishHandler.RegisterRoutes(apiGroup)

	return router
}
