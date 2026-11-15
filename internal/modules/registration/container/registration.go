package container

import (
	"main/internal/modules/registration/controller"
	"main/internal/modules/registration/factory"

	"gorm.io/gorm"
)

type RegistrationContainer struct {
	Controller *controller.RegistrationController
}

func NewRegistrationContainer(db *gorm.DB) *RegistrationContainer {
	return &RegistrationContainer{
		Controller: factory.NewRegistrationController(db),
	}
}
