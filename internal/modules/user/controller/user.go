package controller

import (
	"context"
	"errors"
	"net/http"

	"main/internal/modules/user/domain"
	"main/internal/modules/user/usecase"
	shared "main/internal/shared/pagination"
	apperrors "main/internal/shared/errors"


	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UserUseCaseI is what the controller calls on the user usecase.
type UserUseCaseI interface {
	GetUser(ctx context.Context, page pagination.Pagination) ([]domain.User, error)
	GetUserById(ctx context.Context, id uuid.UUID) (*domain.User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, input usecase.UpdateUserInput) (*domain.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
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

func (c *UserController) GetUser(ctx *gin.Context) {
	var page shared.Pagination
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

	users, err := c.uc.GetUser(ctx.Request.Context(), page)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengambil data user"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": users,
		"meta": gin.H{"page": page.Page, "limit": page.Limit},
	})
}

func (c *UserController) GetUserById(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "id harus berupa UUID"})
		return
	}

	user, err := c.uc.GetUserById(ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, apperrors.ErrUserNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "user tidak ditemukan"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengambil data user"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": user})
}

func (c *UserController) UpdateUser(ctx *gin.Context) {
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

	user, err := c.uc.UpdateUser(ctx.Request.Context(), id, usecase.UpdateUserInput{
		Name:  req.Name,
		Email: req.Email,
	})
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrUserNotFound):
			ctx.JSON(http.StatusNotFound, gin.H{"error": "user tidak ditemukan"})
		case errors.Is(err, apperrors.ErrEmailTaken):
			ctx.JSON(http.StatusConflict, gin.H{"error": "email sudah dipakai"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengubah user"})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": user})
}

func (c *UserController) DeleteUser(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "id harus berupa UUID"})
		return
	}

	if err := c.uc.DeleteUser(ctx.Request.Context(), id); err != nil {
		if errors.Is(err, apperrors.ErrUserNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "user tidak ditemukan"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "gagal menghapus user"})
		return
	}

	ctx.Status(http.StatusNoContent)
}