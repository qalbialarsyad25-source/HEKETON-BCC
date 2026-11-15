package controller

import (
	"context"
	"errors"
	"net/http"

	"main/internal/modules/announcement/domain"
	"main/internal/modules/announcement/dto"
	"main/internal/modules/announcement/usecase"
	"main/internal/shared/pagination"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AnnouncementUseCaseI interface {
	CreateAnnouncement(ctx context.Context, userID uuid.UUID, title string, body string, eventID uuid.UUID) (*domain.Announcement, error)
	GetAnnouncement(ctx context.Context, page pagination.Pagination) ([]domain.Announcement, error)
	GetAnnouncementById(ctx context.Context, id uuid.UUID) (*domain.Announcement, error)
	UpdateAnnouncement(ctx context.Context, userID uuid.UUID, id uuid.UUID, input usecase.UpdateAnnouncementInput) (*domain.Announcement, error)
	DeleteAnnouncement(ctx context.Context, userID uuid.UUID, id uuid.UUID) error
}

type AnnouncementController struct {
	uc AnnouncementUseCaseI
}

func NewAnnouncementController(uc AnnouncementUseCaseI) *AnnouncementController {
	return &AnnouncementController{uc: uc}
}

type updateAnnouncementRequest struct {
	Body string `json:"body" binding:"required"`
}

func (c *AnnouncementController) CreateAnnouncement(ctx *gin.Context) {
	userID, err := getUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user tidak terautentikasi"})
		return
	}

	var req dto.CreateAnnouncementRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	announcement, err := c.uc.CreateAnnouncement(
		ctx.Request.Context(),
		userID,
		req.Title,
		req.Body,
		req.EventID,
	)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrAnnouncementForbidden):
			ctx.JSON(http.StatusForbidden, gin.H{"error": "kamu bukan anggota internal event ini"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "gagal membuat announcement"})
		}
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"data": dto.DomainToResponse(*announcement),
	})
}

func (c *AnnouncementController) GetAnnouncement(ctx *gin.Context) {
	var page pagination.Pagination
	if err := ctx.ShouldBindQuery(&page); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "query tidak valid"})
		return
	}

	if page.Page < 1 {
		page.Page = 1
	}
	if page.Limit < 1 {
		page.Limit = 10
	}
	if page.Limit > 100 {
		page.Limit = 100
	}
	page.Offset = (page.Page - 1) * page.Limit

	announcements, err := c.uc.GetAnnouncement(ctx.Request.Context(), page)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengambil data announcement"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": announcements,
		"meta": gin.H{"page": page.Page, "limit": page.Limit},
	})
}

func (c *AnnouncementController) GetAnnouncementByID(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "id harus berupa UUID"})
		return
	}

	announcement, err := c.uc.GetAnnouncementById(ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrAnnouncementNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "announcement tidak ditemukan"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengambil data announcement"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": announcement})
}

func (c *AnnouncementController) UpdateAnnouncement(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "id harus berupa UUID"})
		return
	}

	userID, err := getUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user tidak terautentikasi"})
		return
	}

	var req updateAnnouncementRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	announcement, err := c.uc.UpdateAnnouncement(
		ctx.Request.Context(),
		userID,
		id,
		usecase.UpdateAnnouncementInput{
			Body: req.Body,
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrAnnouncementNotFound):
			ctx.JSON(http.StatusNotFound, gin.H{"error": "announcement tidak ditemukan"})
		case errors.Is(err, domain.ErrAnnouncementForbidden):
			ctx.JSON(http.StatusForbidden, gin.H{"error": "kamu bukan anggota internal event ini"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengubah announcement"})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": announcement})
}

func (c *AnnouncementController) DeleteAnnouncement(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "id harus berupa UUID"})
		return
	}

	userID, err := getUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user tidak terautentikasi"})
		return
	}

	if err := c.uc.DeleteAnnouncement(ctx.Request.Context(), userID, id); err != nil {
		switch {
		case errors.Is(err, domain.ErrAnnouncementNotFound):
			ctx.JSON(http.StatusNotFound, gin.H{"error": "announcement tidak ditemukan"})
		case errors.Is(err, domain.ErrAnnouncementForbidden):
			ctx.JSON(http.StatusForbidden, gin.H{"error": "kamu bukan anggota internal event ini"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "gagal menghapus announcement"})
		}
		return
	}

	ctx.Status(http.StatusNoContent)
}

func getUserID(ctx *gin.Context) (uuid.UUID, error) {
	userID, exists := ctx.Get("userID")
	if !exists {
		return uuid.Nil, errors.New("user ID tidak ditemukan")
	}

	id, ok := userID.(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("user ID tidak valid")
	}

	return id, nil
}
