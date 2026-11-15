package container

import (
	"main/internal/modules/invitation/controller"
	"main/internal/modules/invitation/repository"
	"main/internal/modules/invitation/routes"
	"main/internal/modules/invitation/usecase"
	"main/internal/shared/ports"

	"gorm.io/gorm"
)

type InvitationContainer struct {
	UseCase *usecase.InvitationUseCase
	Routes  *routes.InvitationRoutes
}

func NewInvitationContainer(db *gorm.DB, adapters *ports.Adapters) *InvitationContainer {
	repo := repository.NewInvitationRepository(db)
	uc := usecase.NewInvitationUseCase(repo)
	ctrl := controller.NewInvitationController(uc)
	r := routes.NewInvitationRoutes(ctrl, adapters.Logger)

	return &InvitationContainer{
		UseCase: uc,
		Routes:  r,
	}
}
