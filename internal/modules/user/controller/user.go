package controller

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"main/internal/modules/user/domain"
	"main/internal/modules/user/usecase"
	"main/internal/shared/pagination"
	"main/internal/modules/user/dto"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UserUseCaseI is what the controller calls on the user usecase.
type UserUseCaseI interface {
	List(ctx context.Context, page pagination.Pagination) (*[]domain.User, int, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	Update(ctx context.Context, id uuid.UUID, input usecase.UpdateUserInput) (*domain.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type UserController struct {
	uc UserUseCaseI
}

func NewUserController(uc UserUseCaseI) *UserController {
	return &UserController{uc: uc}
}

type updateUserRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

func (c *UserController) List(ctx *gin.Context) {
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

	users, total, err := c.uc.List(ctx.Request.Context(), page)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengambil data user"})
		return
	}

	paginationMeta := pagination.GetPaginationResponse(page, total)

	ctx.JSON(http.StatusOK, gin.H{
		"data": users,
		"meta": gin.H{"pagination": paginationMeta},
	})
}

func (c *UserController) GetMe(ctx *gin.Context) {
	// Pastikan key "user_id" ini sama dengan yang di-set di middleware.Authenticate
	rawID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, err := uuid.Parse(fmt.Sprint(rawID))
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id in token"})
		return
	}

	user, err := c.uc.GetByID(ctx.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "user tidak ditemukan"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengambil data user"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": dto.DomainToResponse(*user),
	})
}

func (c *UserController) GetByID(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "id harus berupa UUID"})
		return
	}

	user, err := c.uc.GetByID(ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "user tidak ditemukan"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengambil data user"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": user})
}

func (c *UserController) Update(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "id harus berupa UUID"})
		return
	}

	var req updateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := c.uc.Update(ctx.Request.Context(), id, usecase.UpdateUserInput{
		Name:  req.Name,
		Email: req.Email,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			ctx.JSON(http.StatusNotFound, gin.H{"error": "user tidak ditemukan"})
		case errors.Is(err, domain.ErrEmailTaken):
			ctx.JSON(http.StatusConflict, gin.H{"error": "email sudah dipakai"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengubah user"})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": user})
}

func (c *UserController) Delete(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "id harus berupa UUID"})
		return
	}

	if err := c.uc.Delete(ctx.Request.Context(), id); err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "user tidak ditemukan"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "gagal menghapus user"})
		return
	}

	ctx.Status(http.StatusNoContent)
}
