package mapper

import (
	"main/internal/infrastructure/database"
	"main/internal/modules/user/domain"
)

func ToDomainUser(u *database.User) *domain.User {
	if u == nil {
		return nil
	}

	return &domain.User{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Password:  u.Password,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func ToDatabaseUser(u *domain.User) *database.User {
	if u == nil {
		return nil
	}

	return &database.User{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Password:  u.Password,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
