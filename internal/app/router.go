package app

import (
	"fmt"
	"net/http"

	"main/internal/config"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRouter initializes a fresh Gin router without custom middleware or module containers.
func SetupRouter(cfg *config.Config, db *gorm.DB) *gin.Engine {
	switch cfg.App.Mode {
	case config.ModeProduction:
		gin.SetMode(gin.ReleaseMode)
	case config.ModeTest:
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()
	router.HandleMethodNotAllowed = true
	router.RedirectTrailingSlash = false

	if err := router.SetTrustedProxies(cfg.HTTP.TrustedProxies); err != nil {
		panic(fmt.Sprintf("TRUSTED_PROXIES: %v", err))
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := router.Group("/api/v1")
	{
		// Register routes directly here
		_ = api
	}

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "route not found"})
	})

	router.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
	})

	return router
}
