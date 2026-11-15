package routes

import (
	"main/internal/modules/registration/controller"

	"github.com/gin-gonic/gin"
)

func RegisterRegistrationRoutes(
	r *gin.RouterGroup,
	authMiddleware gin.HandlerFunc,
	ctrl *controller.RegistrationController,
) {
	protected := r.Group("", authMiddleware)
	{
		// Event-scoped registration endpoint
		protected.POST("/events/:id/registration", ctrl.RegisterEvent)

		// Registration collection & user-scoped endpoints (/my registered before /:id)
		registrations := protected.Group("/registrations")
		{
			registrations.GET("/my", ctrl.GetMyRegistrations)
			registrations.GET("", ctrl.GetRegistrations)
			registrations.POST("/:id/cancel", ctrl.CancelRegistration)
			registrations.PUT("/:id", ctrl.UpdateRegistration)
			registrations.DELETE("/:id", ctrl.DeleteRegistration)
			registrations.POST("/:id/check-in", ctrl.CheckInRegistration)
		}
	}
}
