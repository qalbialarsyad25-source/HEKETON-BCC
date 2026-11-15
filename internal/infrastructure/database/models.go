package database

import (
	"time"

	"github.com/google/uuid"
)

const (
	EventStatusUpcoming = "UPCOMING"
	EventStatusOngoing  = "ONGOING"
	EventStatusFinished = "FINISHED"

	SessionStatusUpcoming = "UPCOMING"
	SessionStatusOngoing  = "ONGOING"
	SessionStatusFinished = "FINISHED"

	RoleCoordinator = "COORDINATOR"
	RoleStaff       = "STAFF"

	RegistrationStatusRegistered = "REGISTERED"
	RegistrationStatusCheckedIn  = "CHECKEDIN"
	RegistrationStatusCancelled  = "CANCELLED"

	InvitationStatusPending   = "PENDING"
	InvitationStatusAccepted  = "ACCEPTED"
	InvitationStatusDeclined  = "DECLINED"
	InvitationStatusCancelled = "CANCELLED"
)

type User struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Email     string    `gorm:"not null;uniqueIndex:idx_users_email"`
	Name      string    `gorm:"not null"`
	Password  string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;default:CURRENT_TIMESTAMP"`
}

type Event struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Title          string    `gorm:"not null"`
	Description    string    `gorm:"not null"`
	Date           time.Time `gorm:"not null"`
	Location       string    `gorm:"not null"`
	AvailableQuota int       `gorm:"not null"`
	MaxQuota       int       `gorm:"not null"`
	Status         string    `gorm:"not null;default:UPCOMING;index:idx_events_status"`
	CreatedAt      time.Time `gorm:"autoCreateTime;default:CURRENT_TIMESTAMP"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime;default:CURRENT_TIMESTAMP"`
}

type Announcement struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Title     string    `gorm:"not null"`
	Body      string    `gorm:"not null"`
	EventID   uuid.UUID `gorm:"type:uuid;not null;index:idx_announcements_event_id"`
	SenderID  uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;default:CURRENT_TIMESTAMP"`

	Event  Event `gorm:"foreignKey:EventID;constraint:OnDelete:CASCADE"`
	Sender User  `gorm:"foreignKey:SenderID;constraint:OnDelete:CASCADE"`
}

type Invitation struct {
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	EventID    uuid.UUID `gorm:"type:uuid;not null;index:idx_invitations_event_id"`
	SenderID   uuid.UUID `gorm:"type:uuid;not null"`
	ReceiverID uuid.UUID `gorm:"type:uuid;not null;index:idx_invitations_receiver_id"`
	Role       string    `gorm:"not null"`
	Status     string    `gorm:"not null;default:PENDING"`
	CreatedAt  time.Time `gorm:"autoCreateTime;default:CURRENT_TIMESTAMP"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime;default:CURRENT_TIMESTAMP"`

	Event    Event `gorm:"foreignKey:EventID;constraint:OnDelete:CASCADE"`
	Sender   User  `gorm:"foreignKey:SenderID;constraint:OnDelete:CASCADE"`
	Receiver User  `gorm:"foreignKey:ReceiverID;constraint:OnDelete:CASCADE"`
}

type Registration struct {
	ID        string `gorm:"type:uuid;primaryKey"`
	EventID   string `gorm:"type:uuid;not null;uniqueIndex:idx_event_user_reg"`
	UserID    string `gorm:"type:uuid;not null;uniqueIndex:idx_event_user_reg"`
	Status    string `gorm:"type:varchar(20);not null;default:'REGISTERED'"`
	User      User   `gorm:"foreignKey:UserID"`
	Event     Event  `gorm:"foreignKey:EventID"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
type Committee struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	EventID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_committees_event_user"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_committees_event_user"`
	Role      string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;default:CURRENT_TIMESTAMP"`

	Event Event `gorm:"foreignKey:EventID;constraint:OnDelete:CASCADE"`
	User  User  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

type Session struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	EventID   uuid.UUID `gorm:"type:uuid;not null;index:idx_sessions_event_id"`
	Title     string    `gorm:"not null"`
	Time      time.Time `gorm:"not null"`
	Speaker   string    `gorm:"not null"`
	Location  string    `gorm:"not null"`
	Status    string    `gorm:"not null;default:UPCOMING"`
	CreatedAt time.Time `gorm:"autoCreateTime;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;default:CURRENT_TIMESTAMP"`

	Event Event `gorm:"foreignKey:EventID;constraint:OnDelete:CASCADE"`
}

type RefreshToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	ExpiresAt time.Time `gorm:"not null"`
	RevokedAt *time.Time
	CreatedAt time.Time `gorm:"autoCreateTime;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;default:CURRENT_TIMESTAMP"`
	User      User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func Models() []any {
	return []any{
		&User{},
		&Event{},
		&Announcement{},
		&Invitation{},
		&Registration{},
		&Committee{},
		&Session{},
		&RefreshToken{},
	}
}
