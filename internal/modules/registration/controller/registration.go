package controller

import (
	"errors"
	"main/internal/modules/registration/domain"
	"main/internal/modules/registration/dto"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type RegistrationController struct {
	uc domain.Usecase
}

func NewRegistrationController(uc domain.Usecase) *RegistrationController {
	return &RegistrationController{uc: uc}
}

type paginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

func buildMeta(page, limit int, total int64) paginationMeta {
	totalPages := 0
	if limit > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}
	return paginationMeta{
		Page:       page,
		Limit:      limit,
		TotalItems: total,
		TotalPages: totalPages,
	}
}

func parsePagination(c *gin.Context) (int, int) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}
	return page, limit
}

// Adjust key lookup ("user_id" / "userID") to match internal/http/middleware/authenticate.go
func extractUserID(c *gin.Context) (string, bool) {
	for _, key := range []string{"user_id", "userID", "sub"} {
		if val, exists := c.Get(key); exists {
			if id, ok := val.(string); ok && id != "" {
				return id, true
			}
		}
	}
	return "", false
}

func mapErrorStatus(err error) int {
	switch {
	case errors.Is(err, domain.ErrRegistrationNotFound),
		errors.Is(err, domain.ErrEventNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, domain.ErrAlreadyRegistered),
		errors.Is(err, domain.ErrAlreadyCancelled),
		errors.Is(err, domain.ErrAlreadyCheckedIn),
		errors.Is(err, domain.ErrCannotCancelCheckedIn),
		errors.Is(err, domain.ErrCannotCheckInCancelled),
		errors.Is(err, domain.ErrEventQuotaFull):
		return http.StatusConflict
	case errors.Is(err, domain.ErrEventNotOpen),
		errors.Is(err, domain.ErrInvalidStatus):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func respondError(c *gin.Context, err error) {
	status := mapErrorStatus(err)
	c.JSON(status, gin.H{"error": err.Error()})
}

// POST /events/:id/registration
func (ctrl *RegistrationController) RegisterEvent(c *gin.Context) {
	userID, ok := extractUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	eventID := c.Param("id")
	if eventID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "event id is required"})
		return
	}

	reg, err := ctrl.uc.RegisterEvent(c.Request.Context(), eventID, userID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": dto.ToRegistrationResponse(reg),
	})
}

// GET /registrations/my
func (ctrl *RegistrationController) GetMyRegistrations(c *gin.Context) {
	userID, ok := extractUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	page, limit := parsePagination(c)
	items, total, err := ctrl.uc.GetMyRegistrations(c.Request.Context(), userID, page, limit)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": dto.ToRegistrationListResponse(items),
		"meta": buildMeta(page, limit, total),
	})
}

// GET /registrations
func (ctrl *RegistrationController) GetRegistrations(c *gin.Context) {
	page, limit := parsePagination(c)
	eventID := c.Query("event_id")

	items, total, err := ctrl.uc.GetRegistrations(c.Request.Context(), eventID, page, limit)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": dto.ToRegistrationListResponse(items),
		"meta": buildMeta(page, limit, total),
	})
}

// POST /registrations/:id/cancel
func (ctrl *RegistrationController) CancelRegistration(c *gin.Context) {
	userID, ok := extractUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	regID := c.Param("id")
	reg, err := ctrl.uc.CancelRegistration(c.Request.Context(), regID, userID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": dto.ToRegistrationResponse(reg),
	})
}

// PUT /registrations/:id
func (ctrl *RegistrationController) UpdateRegistration(c *gin.Context) {
	userID, ok := extractUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.UpdateRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	regID := c.Param("id")
	reg, err := ctrl.uc.UpdateRegistration(c.Request.Context(), regID, userID, domain.Status(req.Status))
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": dto.ToRegistrationResponse(reg),
	})
}

// DELETE /registrations/:id
func (ctrl *RegistrationController) DeleteRegistration(c *gin.Context) {
	userID, ok := extractUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	regID := c.Param("id")
	if err := ctrl.uc.DeleteRegistration(c.Request.Context(), regID, userID); err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{},
	})
}

// POST /registrations/:id/check-in
func (ctrl *RegistrationController) CheckInRegistration(c *gin.Context) {
	userID, ok := extractUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	regID := c.Param("id")
	reg, err := ctrl.uc.CheckInRegistration(c.Request.Context(), regID, userID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": dto.ToRegistrationResponse(reg),
	})
}
