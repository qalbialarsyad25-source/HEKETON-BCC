package routes

import (
	"main/internal/http/middleware"
	"main/internal/modules/announcement/controller"
	"main/internal/shared/ports/token"
	ws "main/internal/infrastructure/websocket"

	"github.com/gin-gonic/gin"
)

type AnnouncementRoutes struct {
	ctrl      *controller.AnnouncementController
	tokens    token.TokenService
	wsHandler *ws.Handler
}

func NewAnnouncementRoutes(
	ctrl *controller.AnnouncementController,
	tokens token.TokenService,
	wsHandler *ws.Handler,
) *AnnouncementRoutes {
	return &AnnouncementRoutes{
		ctrl:      ctrl,
		tokens:    tokens,
		wsHandler: wsHandler,
	}
}

func (r AnnouncementRoutes) RegisterRoutes(rg *gin.RouterGroup) {
	router := rg.Group("/announcements")
	router.Use(middleware.Authenticate(r.tokens))

	router.GET(
		"",
		middleware.HandlePagination(10, 100),
		r.ctrl.GetAnnouncement,
	)
	router.POST("", r.ctrl.CreateAnnouncement)
	router.GET("/:id", r.ctrl.GetAnnouncementByID)
	router.PUT("/:id", r.ctrl.UpdateAnnouncement)
	router.DELETE("/:id", r.ctrl.DeleteAnnouncement)

	// WebSocket
	router.GET("/ws", r.wsHandler.ServeWebSocket)
}