package controller

import (
	"context"
	"net/http"

	"main/internal/http/response"
	"main/internal/http/validator"
	"main/internal/modules/invitation/domain"
	"main/internal/modules/invitation/dto"
	"main/internal/modules/invitation/usecase"
	apperrors "main/internal/shared/errors"
	"main/internal/shared/pagination"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// userIDContextKey is the key the auth middleware stores the signed-in user's id under.
const userIDContextKey = "userID"

var errUnauthenticated = apperrors.Unauthorized("UNAUTHENTICATED", "authentication is required")

// InvitationUseCaseI is what the controller calls on the invitation usecase.
type InvitationUseCaseI interface {
	Create(ctx context.Context, input usecase.CreateInvitationInput) (*domain.Invitation, error)
	GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.Invitation, error)
	List(ctx context.Context, userID uuid.UUID, eventID *uuid.UUID, params pagination.Params) (*pagination.Result[domain.Invitation], error)
	ListMine(ctx context.Context, userID uuid.UUID, params pagination.Params) (*pagination.Result[domain.Invitation], error)
	UpdateRole(ctx context.Context, input usecase.UpdateInvitationInput) (*domain.Invitation, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
	Accept(ctx context.Context, id, userID uuid.UUID) (*domain.Invitation, error)
	Decline(ctx context.Context, id, userID uuid.UUID) (*domain.Invitation, error)
	Cancel(ctx context.Context, id, userID uuid.UUID) (*domain.Invitation, error)
}

type InvitationController struct {
	uc InvitationUseCaseI
}

func NewInvitationController(uc InvitationUseCaseI) *InvitationController {
	return &InvitationController{uc: uc}
}

// Create handles POST /invitations.
func (ctrl InvitationController) Create(c *gin.Context) {
	userID, err := currentUserID(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	req, err := validator.BindJSON[dto.CreateInvitationRequest](c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	eventID, err := parseEventID(req.EventID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	receiverID, err := uuid.Parse(req.ReceiverID)
	if err != nil {
		_ = c.Error(domain.ErrInvalidReceiverID.Wrap(err))
		return
	}

	ctrl.create(c, usecase.CreateInvitationInput{
		SenderID:   userID,
		ReceiverID: receiverID,
		EventID:    eventID,
		Role:       domain.Role(req.Role),
	})
}

// InviteUser handles POST /users/:id/invitations. The invited user is the :id in the path.
func (ctrl InvitationController) InviteUser(c *gin.Context) {
	userID, err := currentUserID(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	receiverID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(domain.ErrReceiverNotFound.Wrap(err))
		return
	}

	req, err := validator.BindJSON[dto.InviteUserRequest](c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	eventID, err := parseEventID(req.EventID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	ctrl.create(c, usecase.CreateInvitationInput{
		SenderID:   userID,
		ReceiverID: receiverID,
		EventID:    eventID,
		Role:       domain.Role(req.Role),
	})
}

// List handles GET /invitations: the invitations of the events the user coordinates.
// ?event_id= narrows it to one event.
func (ctrl InvitationController) List(c *gin.Context) {
	userID, err := currentUserID(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	var eventID *uuid.UUID
	if raw := c.Query("event_id"); raw != "" {
		parsed, err := parseEventID(raw)
		if err != nil {
			_ = c.Error(err)
			return
		}

		eventID = &parsed
	}

	params := pagination.ParamsFromQuery(c.Query("page"), c.Query("limit"))

	ctx := c.Request.Context()
	result, err := ctrl.uc.List(ctx, userID, eventID, params)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, dto.InvitationsToListResponse(result))
}

// ListMine handles GET /invitations/my: the invitations sent to the signed-in user.
func (ctrl InvitationController) ListMine(c *gin.Context) {
	userID, err := currentUserID(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	params := pagination.ParamsFromQuery(c.Query("page"), c.Query("limit"))

	ctx := c.Request.Context()
	result, err := ctrl.uc.ListMine(ctx, userID, params)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, dto.InvitationsToListResponse(result))
}

func (ctrl InvitationController) Get(c *gin.Context) {
	userID, err := currentUserID(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	id, err := invitationIDFromPath(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	ctx := c.Request.Context()
	invitation, err := ctrl.uc.GetByID(ctx, id, userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	writeInvitation(c, http.StatusOK, *invitation)
}

// Update handles PUT /invitation/:id, which changes the offered role.
func (ctrl InvitationController) Update(c *gin.Context) {
	userID, err := currentUserID(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	id, err := invitationIDFromPath(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	req, err := validator.BindJSON[dto.UpdateInvitationRequest](c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	input := usecase.UpdateInvitationInput{
		ID:     id,
		UserID: userID,
		Role:   domain.Role(req.Role),
	}

	ctx := c.Request.Context()
	invitation, err := ctrl.uc.UpdateRole(ctx, input)
	if err != nil {
		_ = c.Error(err)
		return
	}

	writeInvitation(c, http.StatusOK, *invitation)
}

func (ctrl InvitationController) Delete(c *gin.Context) {
	userID, err := currentUserID(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	id, err := invitationIDFromPath(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	ctx := c.Request.Context()
	if err := ctrl.uc.Delete(ctx, id, userID); err != nil {
		_ = c.Error(err)
		return
	}

	body := response.ResponseDTO[struct{}]{
		StatusCode: http.StatusOK,
		Data:       struct{}{},
	}

	response.Success(c, body)
}

// Accept handles POST /invitations/:id/accept.
func (ctrl InvitationController) Accept(c *gin.Context) {
	ctrl.answer(c, ctrl.uc.Accept)
}

// Decline handles POST /invitations/:id/decline.
func (ctrl InvitationController) Decline(c *gin.Context) {
	ctrl.answer(c, ctrl.uc.Decline)
}

// Cancel handles POST /invitations/:id/cancel.
func (ctrl InvitationController) Cancel(c *gin.Context) {
	ctrl.answer(c, ctrl.uc.Cancel)
}

func (ctrl InvitationController) create(c *gin.Context, input usecase.CreateInvitationInput) {
	ctx := c.Request.Context()
	invitation, err := ctrl.uc.Create(ctx, input)
	if err != nil {
		_ = c.Error(err)
		return
	}

	writeInvitation(c, http.StatusCreated, *invitation)
}

// answer runs one of the accept, decline and cancel actions on the invitation in the path.
func (ctrl InvitationController) answer(c *gin.Context, action func(ctx context.Context, id, userID uuid.UUID) (*domain.Invitation, error)) {
	userID, err := currentUserID(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	id, err := invitationIDFromPath(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	ctx := c.Request.Context()
	invitation, err := action(ctx, id, userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	writeInvitation(c, http.StatusOK, *invitation)
}

func writeInvitation(c *gin.Context, status int, invitation domain.Invitation) {
	body := response.ResponseDTO[dto.InvitationResponse]{
		StatusCode: status,
		Data:       *dto.InvitationToResponse(invitation),
	}

	response.Success(c, body)
}

// invitationIDFromPath reads :id. An id that isn't a UUID can't belong to any invitation, so it is a 404.
func invitationIDFromPath(c *gin.Context) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return uuid.Nil, domain.ErrInvitationNotFound.Wrap(err)
	}

	return id, nil
}

func parseEventID(raw string) (uuid.UUID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, domain.ErrInvalidEventID.Wrap(err)
	}

	return id, nil
}

func currentUserID(c *gin.Context) (uuid.UUID, error) {
	value, ok := c.Get(userIDContextKey)
	if !ok {
		return uuid.Nil, errUnauthenticated
	}

	switch id := value.(type) {
	case uuid.UUID:
		return id, nil
	case string:
		parsed, err := uuid.Parse(id)
		if err != nil {
			return uuid.Nil, errUnauthenticated.Wrap(err)
		}
		return parsed, nil
	default:
		return uuid.Nil, errUnauthenticated
	}
}
