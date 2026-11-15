package controller

import (
	"context"
	"net/http"

	"main/internal/http/response"
	"main/internal/http/validator"
	"main/internal/modules/event/domain"
	"main/internal/modules/event/dto"
	"main/internal/modules/event/usecase"
	"main/internal/shared/pagination"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// EventUseCaseI is what the controller calls on the event usecase.
type EventUseCaseI interface {
	Create(ctx context.Context, input usecase.CreateEventInput) (*domain.Event, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Event, error)
	List(ctx context.Context, params pagination.Params) (*pagination.Result[domain.Event], error)
	ListMine(ctx context.Context, userID uuid.UUID, params pagination.Params) (*pagination.Result[domain.Event], error)
	Update(ctx context.Context, input usecase.UpdateEventInput) (*domain.Event, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
}

type EventController struct {
	uc EventUseCaseI
}

func NewEventController(uc EventUseCaseI) *EventController {
	return &EventController{uc: uc}
}

func (ctrl EventController) Create(c *gin.Context) {
	userID, err := currentUserID(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	req, err := validator.BindJSON[dto.CreateEventRequest](c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	input := usecase.CreateEventInput{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Date:        req.Date,
		Location:    req.Location,
		MaxQuota:    req.MaxQuota,
	}

	ctx := c.Request.Context()
	event, err := ctrl.uc.Create(ctx, input)
	if err != nil {
		_ = c.Error(err)
		return
	}

	body := response.ResponseDTO[dto.EventResponse]{
		StatusCode: http.StatusCreated,
		Data:       *dto.EventToResponse(*event),
	}

	response.Success(c, body)
}

func (ctrl EventController) List(c *gin.Context) {
	params := pagination.ParamsFromQuery(c.Query("page"), c.Query("limit"))

	ctx := c.Request.Context()
	result, err := ctrl.uc.List(ctx, params)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, dto.EventsToListResponse(result))
}

func (ctrl EventController) ListMine(c *gin.Context) {
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

	c.JSON(http.StatusOK, dto.EventsToListResponse(result))
}

func (ctrl EventController) Get(c *gin.Context) {
	id, err := eventIDFromPath(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	ctx := c.Request.Context()
	event, err := ctrl.uc.GetByID(ctx, id)
	if err != nil {
		_ = c.Error(err)
		return
	}

	body := response.ResponseDTO[dto.EventResponse]{
		StatusCode: http.StatusOK,
		Data:       *dto.EventToResponse(*event),
	}

	response.Success(c, body)
}

func (ctrl EventController) Update(c *gin.Context) {
	userID, err := currentUserID(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	id, err := eventIDFromPath(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	req, err := validator.BindJSON[dto.UpdateEventRequest](c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	input := usecase.UpdateEventInput{
		ID:          id,
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Date:        req.Date,
		Location:    req.Location,
		MaxQuota:    req.MaxQuota,
		Status:      domain.Status(req.Status),
	}

	ctx := c.Request.Context()
	event, err := ctrl.uc.Update(ctx, input)
	if err != nil {
		_ = c.Error(err)
		return
	}

	body := response.ResponseDTO[dto.EventResponse]{
		StatusCode: http.StatusOK,
		Data:       *dto.EventToResponse(*event),
	}

	response.Success(c, body)
}

func (ctrl EventController) Delete(c *gin.Context) {
	userID, err := currentUserID(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	id, err := eventIDFromPath(c)
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

// eventIDFromPath reads :id. An id that isn't a UUID can't belong to any event, so it is a 404.
func eventIDFromPath(c *gin.Context) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return uuid.Nil, domain.ErrEventNotFound.Wrap(err)
	}

	return id, nil
}
