package mapper

import (
	"main/internal/infrastructure/database"
	"main/internal/modules/announcement/domain"
)

func ToDomainAnnouncement(a *database.Announcement) *domain.Announcement {
	if a == nil {
		return nil
	}

	return &domain.Announcement{
		ID:        a.ID,
		Title:     a.Title,
		Body:      a.Body,
		SenderId:  a.SenderID,
		EventId:   a.EventID,
		CreatedAt: a.CreatedAt,
		UpdateAt:  a.UpdatedAt,
	}
}

func ToDatabaseAnnouncement(a *domain.Announcement) *database.Announcement {
	if a == nil {
		return nil
	}

	return &database.Announcement{
		ID:        a.ID,
		Title:     a.Title,
		Body:      a.Body,
		SenderID:  a.SenderId,
		EventID:   a.EventId,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdateAt,
	}
}

func ToDomainAnnouncementBatch(announcements []database.Announcement) []domain.Announcement {
	result := make([]domain.Announcement, 0, len(announcements))

	for _, a := range announcements {
		result = append(result, *ToDomainAnnouncement(&a))
	}

	return result
}