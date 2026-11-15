package container

import (
	"main/internal/modules/event/controller"
	"main/internal/modules/event/repository"
	"main/internal/modules/event/routes"
	"main/internal/modules/event/usecase"
	"main/internal/shared/ports"

	"gorm.io/gorm"
)

type EventContainer struct {
	UseCase *usecase.EventUseCase
	Routes  *routes.EventRoutes
}

func NewEventContainer(db *gorm.DB, adapters *ports.Adapters) *EventContainer {
	repo := repository.NewEventRepository(db)
	uc := usecase.NewEventUseCase(repo)
	ctrl := controller.NewEventController(uc)
	r := routes.NewEventRoutes(ctrl, adapters.Logger)

	return &EventContainer{
		UseCase: uc,
		Routes:  r,
	}
}
