package controller

import (
	"context"
	"net/http"

	"main/internal/http/middleware"
	"main/internal/http/response"
	"main/internal/http/validator"
	"main/internal/modules/session/domain"
	"main/internal/modules/session/dto"
	"main/internal/modules/session/usecase"
	"main/internal/shared/pagination"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SessionUseCaseI is what the controller calls on the session usecase.
type SessionUseCaseI interface {
	Create(ctx context.Context, input usecase.CreateSessionInput) (*domain.Session, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Session, error)
	List(ctx context.Context, eventID *uuid.UUID, params pagination.Params) (*pagination.Result[domain.Session], error)
	Update(ctx context.Context, input usecase.UpdateSessionInput) (*domain.Session, error)
	Delete(ctx context.Context, id, userID uuid.UUID) (*domain.Session, error)
	EventExists(ctx context.Context, eventID uuid.UUID) (bool, error)
}

// EventStreamer upgrades a request to a WebSocket that follows one event.
type EventStreamer interface {
	Serve(c *gin.Context, eventID uuid.UUID)
}

type SessionController struct {
	uc     SessionUseCaseI
	stream EventStreamer
}

func NewSessionController(uc SessionUseCaseI, stream EventStreamer) *SessionController {
	return &SessionController{uc: uc, stream: stream}
}

// Create handles POST /sessions.
func (ctrl SessionController) Create(c *gin.Context) {
	actor, err := middleware.GetActor(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	req, err := validator.BindJSON[dto.CreateSessionRequest](c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	eventID, err := uuid.Parse(req.EventID)
	if err != nil {
		_ = c.Error(domain.ErrInvalidEventID.Wrap(err))
		return
	}

	input := usecase.CreateSessionInput{
		UserID:   actor.UserID,
		EventID:  eventID,
		Title:    req.Title,
		Time:     req.Time,
		Speaker:  req.Speaker,
		Location: req.Location,
	}

	ctx := c.Request.Context()
	session, err := ctrl.uc.Create(ctx, input)
	if err != nil {
		_ = c.Error(err)
		return
	}

	writeSession(c, http.StatusCreated, *session)
}

// List handles GET /sessions. ?event_id= narrows it to one event.
func (ctrl SessionController) List(c *gin.Context) {
	var eventID *uuid.UUID
	if raw := c.Query("event_id"); raw != "" {
		parsed, err := uuid.Parse(raw)
		if err != nil {
			_ = c.Error(domain.ErrInvalidEventID.Wrap(err))
			return
		}

		eventID = &parsed
	}

	params := pagination.ParamsFromQuery(c.Query("page"), c.Query("limit"))

	ctx := c.Request.Context()
	result, err := ctrl.uc.List(ctx, eventID, params)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, dto.SessionsToListResponse(result))
}

// Get handles GET /session/:id.
func (ctrl SessionController) Get(c *gin.Context) {
	id, err := sessionIDFromPath(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	ctx := c.Request.Context()
	session, err := ctrl.uc.GetByID(ctx, id)
	if err != nil {
		_ = c.Error(err)
		return
	}

	writeSession(c, http.StatusOK, *session)
}

// Update handles PUT /session/:id.
func (ctrl SessionController) Update(c *gin.Context) {
	actor, err := middleware.GetActor(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	id, err := sessionIDFromPath(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	req, err := validator.BindJSON[dto.UpdateSessionRequest](c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	input := usecase.UpdateSessionInput{
		ID:       id,
		UserID:   actor.UserID,
		Title:    req.Title,
		Time:     req.Time,
		Speaker:  req.Speaker,
		Location: req.Location,
		Status:   domain.Status(req.Status),
	}

	ctx := c.Request.Context()
	session, err := ctrl.uc.Update(ctx, input)
	if err != nil {
		_ = c.Error(err)
		return
	}

	writeSession(c, http.StatusOK, *session)
}

// Delete handles DELETE /session/:id and answers with the session that was removed.
func (ctrl SessionController) Delete(c *gin.Context) {
	actor, err := middleware.GetActor(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	id, err := sessionIDFromPath(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	ctx := c.Request.Context()
	session, err := ctrl.uc.Delete(ctx, id, actor.UserID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	writeSession(c, http.StatusOK, *session)
}

// Stream handles GET /ws/events/:eventID/sessions. Every change to a session of the event
// is pushed to the connection as {"type": "session.created|updated|deleted", "data": {...}}.
func (ctrl SessionController) Stream(c *gin.Context) {
	eventID, err := uuid.Parse(c.Param("eventID"))
	if err != nil {
		_ = c.Error(domain.ErrEventNotFound.Wrap(err))
		return
	}

	exists, err := ctrl.uc.EventExists(c.Request.Context(), eventID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	if !exists {
		_ = c.Error(domain.ErrEventNotFound)
		return
	}

	ctrl.stream.Serve(c, eventID)
}

func writeSession(c *gin.Context, status int, session domain.Session) {
	body := response.ResponseDTO[dto.SessionResponse]{
		StatusCode: status,
		Data:       *dto.SessionToResponse(session),
	}

	response.Success(c, body)
}

// sessionIDFromPath reads :id. An id that isn't a UUID can't belong to any session, so it is a 404.
func sessionIDFromPath(c *gin.Context) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return uuid.Nil, domain.ErrSessionNotFound.Wrap(err)
	}

	return id, nil
}
