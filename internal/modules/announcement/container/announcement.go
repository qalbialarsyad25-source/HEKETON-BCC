package container

import (
	ws "main/internal/infrastructure/websocket"
	"main/internal/modules/announcement/controller"
	"main/internal/modules/announcement/repository"
	"main/internal/modules/announcement/routes"
	"main/internal/modules/announcement/usecase"
	"main/internal/shared/ports"

	"gorm.io/gorm"
)

type AnnouncementContainer struct {
	UseCase *usecase.AnnouncementUseCase
	Routes  *routes.AnnouncementRoutes
}

func NewAnnouncementContainer(db *gorm.DB, adapters *ports.Adapters, membership usecase.EventMembershipRepositoryI, publisher usecase.AnnouncementPublisher, ws *ws.Handler) *AnnouncementContainer {
	repo := repository.NewAnnouncementRepository(db)
	uc := usecase.NewAnnouncementUseCase(repo, membership, publisher)
	ctrl := controller.NewAnnouncementController(uc)
	r := routes.NewAnnouncementRoutes(ctrl, adapters.TokenService, ws)

	return &AnnouncementContainer{
		UseCase: &uc,
		Routes:  r,
	}
}
