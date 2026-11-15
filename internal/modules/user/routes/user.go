package routes

import (
	"main/internal/http/middleware"
	"main/internal/modules/user/controller"
	"main/internal/shared/ports/token"

	"github.com/gin-gonic/gin"
)

type UserRoutes struct {
	ctrl   *controller.UserController
	tokens token.TokenService
}

func NewUserRoutes(ctrl *controller.UserController, tokens token.TokenService) *UserRoutes {
	return &UserRoutes{ctrl: ctrl, tokens: tokens}
}

func (r UserRoutes) RegisterRoutes(rg *gin.RouterGroup) {
	router := rg.Group("/users")
	router.Use(middleware.Authenticate(r.tokens))

	router.GET(
		"",
		middleware.HandlePagination(10, 100),
		r.ctrl.List,
	)
	router.GET("/:id", r.ctrl.GetByID)
	router.PUT("/:id", r.ctrl.Update)
	router.DELETE("/:id", r.ctrl.Delete)
}
