package container

import (
	ws "main/internal/infrastructure/websocket"
	"main/internal/modules/session/controller"
	"main/internal/modules/session/repository"
	"main/internal/modules/session/routes"
	"main/internal/modules/session/usecase"
	"main/internal/shared/ports"

	"gorm.io/gorm"
)

type SessionContainer struct {
	UseCase *usecase.SessionUseCase
	Routes  *routes.SessionRoutes
}

func NewSessionContainer(db *gorm.DB, adapters *ports.Adapters) *SessionContainer {
	// Sessions get their own hub, so a session change never reaches the clients of another feed.
	hub := ws.NewHub()

	repo := repository.NewSessionRepository(db)
	uc := usecase.NewSessionUseCase(repo, controller.NewBroadcaster(hub))
	ctrl := controller.NewSessionController(uc, ws.NewHandler(hub, nil))
	r := routes.NewSessionRoutes(ctrl, adapters.TokenService)

	return &SessionContainer{
		UseCase: uc,
		Routes:  r,
	}
}
