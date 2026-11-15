package factory

import (
	"main/internal/modules/registration/controller"
	"main/internal/modules/registration/repository"
	"main/internal/modules/registration/usecase"

	"gorm.io/gorm"
)

func NewRegistrationController(db *gorm.DB) *controller.RegistrationController {
	repo := repository.NewRegistrationRepository(db)
	authz := repository.NewCommitteeAuthorizer(db)
	uc := usecase.NewRegistrationUsecase(repo, authz)
	return controller.NewRegistrationController(uc)
}
