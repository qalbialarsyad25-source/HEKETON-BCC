package routes

import (
	"main/internal/modules/invitation/controller"
	"main/internal/shared/ports/logger"

	"github.com/gin-gonic/gin"
)

type InvitationRoutes struct {
	ctrl   *controller.InvitationController
	logger logger.Logger
}

func NewInvitationRoutes(ctrl *controller.InvitationController, logger logger.Logger) *InvitationRoutes {
	return &InvitationRoutes{ctrl: ctrl, logger: logger}
}

// RegisterRoutes mounts the invitation routes. Every route needs a signed-in user, so mount
// it on the group that already runs the auth middleware. The paths follow the API spec, which
// mixes /invitations (collection and actions) with /invitation/:id (one invitation).
// Who may do what is checked in the usecase: coordinators send, edit, cancel and delete;
// the receiver accepts or declines.
func (r InvitationRoutes) RegisterRoutes(rg *gin.RouterGroup) {
	invitations := rg.Group("/invitations")
	invitations.GET("", r.ctrl.List)
	invitations.POST("", r.ctrl.Create)
	invitations.GET("/my", r.ctrl.ListMine)
	invitations.POST("/:id/accept", r.ctrl.Accept)
	invitations.POST("/:id/decline", r.ctrl.Decline)
	invitations.POST("/:id/cancel", r.ctrl.Cancel)

	invitation := rg.Group("/invitation")
	invitation.GET("/:id", r.ctrl.Get)
	invitation.PUT("/:id", r.ctrl.Update)
	invitation.DELETE("/:id", r.ctrl.Delete)

	rg.POST("/users/:id/invitations", r.ctrl.InviteUser)
}
