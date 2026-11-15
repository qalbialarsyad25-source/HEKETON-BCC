package container

import (
	"main/internal/config"
	"main/internal/modules/auth/controller"
	"main/internal/modules/auth/repository"
	"main/internal/modules/auth/routes"
	"main/internal/modules/auth/usecase"
	"main/internal/shared/ports"

	"gorm.io/gorm"
)

type AuthContainer struct {
	UseCase *usecase.AuthUseCase
	Routes  *routes.AuthRoutes
}

func NewAuthContainer(db *gorm.DB, adapters *ports.Adapters, cfg *config.Config, users usecase.UserService) *AuthContainer {
	repo := repository.NewAuthRepository(db)
	uc := usecase.NewAuthUseCase(repo, users, adapters.TokenService)
	ctrl := controller.NewAuthController(uc, controller.CookieOptions{
		Domain: cfg.HTTP.CookieDomain,
		Secure: cfg.HTTP.CookieSecure,
	})
	r := routes.NewAuthRoutes(ctrl, adapters.Logger)

	return &AuthContainer{
		UseCase: uc,
		Routes:  r,
	}
}
