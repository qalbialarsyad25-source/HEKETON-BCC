package mapper

import (
	"main/internal/infrastructure/database"
	"main/internal/modules/registration/domain"
)

func ToDomain(m *database.Registration) *domain.Registration {
	if m == nil {
		return nil
	}
	return &domain.Registration{
		ID:        m.ID,
		EventID:   m.EventID,
		UserID:    m.UserID,
		Status:    domain.Status(m.Status),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		User: domain.UserSummary{
			ID:    m.UserID,
			Email: m.User.Email,
			Name:  m.User.Name,
		},
	}
}

func ToDomainList(models []database.Registration) []domain.Registration {
	out := make([]domain.Registration, 0, len(models))
	for i := range models {
		out = append(out, *ToDomain(&models[i]))
	}
	return out
}
