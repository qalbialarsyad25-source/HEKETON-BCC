package routes

import (
	"main/internal/http/middleware"
	"main/internal/modules/session/controller"
	"main/internal/shared/ports/token"

	"github.com/gin-gonic/gin"
)

type SessionRoutes struct {
	ctrl   *controller.SessionController
	tokens token.TokenService
}

func NewSessionRoutes(ctrl *controller.SessionController, tokens token.TokenService) *SessionRoutes {
	return &SessionRoutes{ctrl: ctrl, tokens: tokens}
}

// RegisterRoutes mounts the session routes. The paths follow the API spec, which mixes
// /sessions (collection) with /session/:id (one session). Reading is open to any signed-in
// user; creating, updating and deleting are limited to the event's committee, which the
// usecase checks. The WebSocket route takes no token, as the spec says.
func (r SessionRoutes) RegisterRoutes(rg *gin.RouterGroup) {
	sessions := rg.Group("/sessions")
	sessions.Use(middleware.Authenticate(r.tokens))
	sessions.GET("", r.ctrl.List)
	sessions.POST("", r.ctrl.Create)

	session := rg.Group("/session")
	session.Use(middleware.Authenticate(r.tokens))
	session.GET("/:id", r.ctrl.Get)
	session.PUT("/:id", r.ctrl.Update)
	session.DELETE("/:id", r.ctrl.Delete)

	rg.GET("/ws/events/:eventID/sessions", r.ctrl.Stream)
}
