package container

import (
	"main/internal/modules/user/controller"
	"main/internal/modules/user/repository"
	"main/internal/modules/user/routes"
	"main/internal/modules/user/usecase"
	"main/internal/shared/ports"

	"gorm.io/gorm"
)

type UserContainer struct {
	UseCase *usecase.UserUseCase
	Routes  *routes.UserRoutes
}

func NewUserContainer(db *gorm.DB, adapters *ports.Adapters) *UserContainer {
	repo := repository.NewUserRepository(db)
	uc := usecase.NewUserUseCase(repo, adapters.Hasher)
	ctrl := controller.NewUserController(uc)
	r := routes.NewUserRoutes(ctrl, adapters.TokenService)

	return &UserContainer{
		UseCase: &uc,
		Routes:  r,
	}
}
