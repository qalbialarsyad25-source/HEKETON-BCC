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

// NewAuthContainer builds the auth module. users is the user module's usecase: auth reaches
// the users table only through it. adapters.OAuth may be nil (Google sign-in off); tests put
// a fake provider there.
func NewAuthContainer(db *gorm.DB, adapters *ports.Adapters, cfg *config.Config, users usecase.UserService) *AuthContainer {
	repo := repository.NewAuthRepository(db)
	uc := usecase.NewAuthUseCase(repo, users, adapters.TokenService, adapters.OAuth, adapters.Cache)
	ctrl := controller.NewAuthController(uc, controller.CookieOptions{
		Domain: cfg.HTTP.CookieDomain,
		Secure: cfg.HTTP.CookieSecure,
	}, controller.OAuthRedirects{
		Success: cfg.OAuth.SuccessRedirect,
		Failure: cfg.OAuth.FailureRedirect,
	})
	r := routes.NewAuthRoutes(ctrl, adapters.Cache, adapters.Logger, cfg.RateLimit)

	return &AuthContainer{
		UseCase: uc,
		Routes:  r,
	}
}
