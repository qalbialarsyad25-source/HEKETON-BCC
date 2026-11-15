package app

import (
	"fmt"
	"net/http"

	"main/internal/config"
	"main/internal/http/middleware"
	authcontainer "main/internal/modules/auth/container"
	eventcontainer "main/internal/modules/event/container"
	invitationcontainer "main/internal/modules/invitation/container"
	usercontainer "main/internal/modules/user/container"
	"main/internal/shared/ports"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(cfg *config.Config, adapters *ports.Adapters, db *gorm.DB) *gin.Engine {
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

	router.Use(gin.Recovery())
	router.Use(middleware.ErrorHandler())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := router.Group("/api/v1")
	userContainer := usercontainer.NewUserContainer(db, adapters)
	userContainer.Routes.RegisterRoutes(api)

	authContainer := authcontainer.NewAuthContainer(db, adapters, cfg, userContainer.UseCase)
	authContainer.Routes.RegisterRoutes(api)

	eventContainer := eventcontainer.NewEventContainer(db, adapters)
	eventContainer.Routes.RegisterRoutes(api)

	sessionContainer := invitationcontainer.NewInvitationContainer(db, adapters)
	sessionContainer.Routes.RegisterRoutes(api)

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "route not found"})
	})

	router.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
	})

	return router
}
