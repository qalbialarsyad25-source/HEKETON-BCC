package mapper

import (
	"main/internal/infrastructure/database"
	"main/internal/modules/event/domain"
)

func ToDomainEvent(e *database.Event) *domain.Event {
	if e == nil {
		return nil
	}

	return &domain.Event{
		ID:             e.ID,
		Title:          e.Title,
		Description:    e.Description,
		Date:           e.Date,
		Location:       e.Location,
		AvailableQuota: e.AvailableQuota,
		MaxQuota:       e.MaxQuota,
		Status:         domain.Status(e.Status),
		CreatedAt:      e.CreatedAt,
		UpdatedAt:      e.UpdatedAt,
	}
}

func ToDomainEvents(rows []database.Event) []domain.Event {
	events := make([]domain.Event, 0, len(rows))
	for i := range rows {
		events = append(events, *ToDomainEvent(&rows[i]))
	}

	return events
}

func ToDatabaseEvent(e *domain.Event) *database.Event {
	if e == nil {
		return nil
	}

	return &database.Event{
		ID:             e.ID,
		Title:          e.Title,
		Description:    e.Description,
		Date:           e.Date,
		Location:       e.Location,
		AvailableQuota: e.AvailableQuota,
		MaxQuota:       e.MaxQuota,
		Status:         string(e.Status),
		CreatedAt:      e.CreatedAt,
		UpdatedAt:      e.UpdatedAt,
	}
}
