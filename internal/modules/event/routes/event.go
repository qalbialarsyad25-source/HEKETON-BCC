package routes

import (
	"main/internal/modules/event/controller"
	"main/internal/shared/ports/logger"

	"github.com/gin-gonic/gin"
)

type EventRoutes struct {
	ctrl   *controller.EventController
	logger logger.Logger
}

func NewEventRoutes(ctrl *controller.EventController, logger logger.Logger) *EventRoutes {
	return &EventRoutes{ctrl: ctrl, logger: logger}
}

// RegisterRoutes mounts /events. Every route needs a signed-in user, so mount it on the group
// that already runs the auth middleware. Reading is open to any signed-in user; updating and
// deleting are limited to the event's coordinators, which the usecase checks.
func (r EventRoutes) RegisterRoutes(rg *gin.RouterGroup) {
	router := rg.Group("/events")

	router.GET("", r.ctrl.List)
	router.POST("", r.ctrl.Create)
	router.GET("/my", r.ctrl.ListMine)
	router.GET("/:id", r.ctrl.Get)
	router.PUT("/:id", r.ctrl.Update)
	router.DELETE("/:id", r.ctrl.Delete)
}
