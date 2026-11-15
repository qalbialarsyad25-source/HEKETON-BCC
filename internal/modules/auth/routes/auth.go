package routes

import (
	"main/internal/modules/auth/controller"
	"main/internal/shared/ports/logger"

	"github.com/gin-gonic/gin"
)

type AuthRoutes struct {
	ctrl   *controller.AuthController
	logger logger.Logger
}

func NewAuthRoutes(ctrl *controller.AuthController, logger logger.Logger) *AuthRoutes {
	return &AuthRoutes{ctrl: ctrl, logger: logger}
}

// RegisterRoutes mounts /auth. None of these routes take an access token: refresh and logout
// must work after the access token expired. Register, login and the Google start are rate
// limited per client IP. The Google routes are always mounted; with Google sign-in off they
// answer 404 OAUTH_DISABLED.
func (r AuthRoutes) RegisterRoutes(rg *gin.RouterGroup) {
	router := rg.Group("/auth")

	router.POST("/register", r.ctrl.Register)
	router.POST("/login", r.ctrl.Login)
	router.POST("/refresh", r.ctrl.Refresh)
	router.POST("/logout", r.ctrl.Logout)
}
