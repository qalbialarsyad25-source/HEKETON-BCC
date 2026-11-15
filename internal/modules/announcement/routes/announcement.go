package routes

import (
	"main/internal/http/middleware"
	"main/internal/modules/announcement/controller"
	"main/internal/shared/ports/token"

	"github.com/gin-gonic/gin"
)

type AnnouncementRoutes struct {
	ctrl   *controller.AnnouncementController
	tokens token.TokenService
}

func NewAnnouncementRoutes(ctrl *controller.AnnouncementController, tokens token.TokenService) *AnnouncementRoutes {
	return &AnnouncementRoutes{ctrl: ctrl, tokens: tokens}
}

func (r AnnouncementRoutes) RegisterRoutes(rg *gin.RouterGroup) {
	router := rg.Group("/announcements")
	router.Use(middleware.Authenticate(r.tokens))

	router.GET(
		"",
		middleware.HandlePagination(10, 100),
		r.ctrl.GetAnnouncement,
	)
	router.GET("/:id", r.ctrl.GetAnnouncementByID)
	router.PUT("/:id", r.ctrl.UpdateAnnouncement)
	router.DELETE("/:id", r.ctrl.DeleteAnnouncement)
}