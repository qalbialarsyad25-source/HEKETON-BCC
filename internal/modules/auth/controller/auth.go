package controller

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"

	"main/internal/http/response"
	"main/internal/http/validator"
	"main/internal/modules/auth/domain"
	"main/internal/modules/auth/dto"
	"main/internal/modules/auth/usecase"
	apperrors "main/internal/shared/errors"

	"github.com/gin-gonic/gin"
)

// AuthUseCaseI is what the controller calls on the auth usecase.
type AuthUseCaseI interface {
	Register(ctx context.Context, input usecase.RegisterInput) (*domain.Session, error)
	Login(ctx context.Context, input usecase.LoginInput) (*domain.Session, error)
	Refresh(ctx context.Context, input usecase.RefreshInput) (*domain.Session, error)
	Logout(ctx context.Context, input usecase.LogoutInput) error
}

// OAuthRedirects are the frontend pages the Google callback sends the browser to.
type OAuthRedirects struct {
	Success string
	Failure string // ?error=<code> is added
}

type AuthController struct {
	uc        AuthUseCaseI
	cookie    CookieOptions
	redirects OAuthRedirects
}

func NewAuthController(uc AuthUseCaseI, cookie CookieOptions, redirects OAuthRedirects) *AuthController {
	return &AuthController{uc: uc, cookie: cookie, redirects: redirects}
}

func (ctrl AuthController) Register(c *gin.Context) {
	req, err := validator.BindJSON[dto.RegisterRequest](c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	input := usecase.RegisterInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	ctx := c.Request.Context()
	session, err := ctrl.uc.Register(ctx, input)
	if err != nil {
		_ = c.Error(err)
		return
	}

	setRefreshCookie(c, ctrl.cookie, session)

	body := response.ResponseDTO[dto.SessionResponse]{
		StatusCode: http.StatusCreated,
		Data:       *dto.SessionToResponse(*session),
	}

	response.Success(c, body)
}

func (ctrl AuthController) Refresh(c *gin.Context) {
	token, err := readRefreshToken(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	if token == "" {
		_ = c.Error(domain.ErrRefreshTokenMissing)
		return
	}

	ctx := c.Request.Context()
	session, err := ctrl.uc.Refresh(ctx, usecase.RefreshInput{RefreshToken: token})
	if err != nil {
		_ = c.Error(err)
		return
	}

	setRefreshCookie(c, ctrl.cookie, session)

	body := response.ResponseDTO[dto.SessionResponse]{
		StatusCode: http.StatusOK,
		Data:       *dto.SessionToResponse(*session),
	}

	response.Success(c, body)
}

func (ctrl AuthController) Logout(c *gin.Context) {
	clearRefreshCookie(c, ctrl.cookie)

	token, err := readRefreshToken(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	if token != "" {
		ctx := c.Request.Context()
		if err := ctrl.uc.Logout(ctx, usecase.LogoutInput{RefreshToken: token}); err != nil {
			_ = c.Error(err)
			return
		}
	}

	c.Status(http.StatusNoContent)
}

func (ctrl AuthController) Login(c *gin.Context) {
	req, err := validator.BindJSON[dto.LoginRequest](c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	input := usecase.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	}

	ctx := c.Request.Context()
	session, err := ctrl.uc.Login(ctx, input)
	if err != nil {
		_ = c.Error(err)
		return
	}

	setRefreshCookie(c, ctrl.cookie, session)

	body := response.ResponseDTO[dto.SessionResponse]{
		StatusCode: http.StatusOK,
		Data:       *dto.SessionToResponse(*session),
	}

	response.Success(c, body)
}

const (
	// RefreshCookieName is the name of the httpOnly cookie that carries the refresh token.
	RefreshCookieName = "refresh_token"
	// RefreshCookiePath limits the cookie to the auth routes, so it isn't sent with every API call.
	RefreshCookiePath = "/api/v1/auth"
)

// CookieOptions are the deployment-specific parts of the refresh cookie.
type CookieOptions struct {
	Domain string // empty = the API host only
	Secure bool   // HTTPS only; on in production
}

// setRefreshCookie uses http.SetCookie because gin's c.SetCookie can't set SameSite per cookie.
// Lax works because the frontend and the API are on the same site (sub-domains of one domain).
func setRefreshCookie(c *gin.Context, opts CookieOptions, session *domain.Session) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     RefreshCookieName,
		Value:    session.RefreshToken,
		Path:     RefreshCookiePath,
		Domain:   opts.Domain,
		MaxAge:   int(session.RefreshExpiresIn.Seconds()),
		Secure:   opts.Secure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// clearRefreshCookie deletes the cookie. The browser only deletes it when path and domain match.
func clearRefreshCookie(c *gin.Context, opts CookieOptions) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     RefreshCookieName,
		Value:    "",
		Path:     RefreshCookiePath,
		Domain:   opts.Domain,
		MaxAge:   -1,
		Secure:   opts.Secure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// readRefreshToken returns the refresh token from the JSON body ({"refreshToken"}), or else from
// the cookie, or "" when there is none. The body is optional; a body that isn't JSON is an error.
func readRefreshToken(c *gin.Context) (string, error) {
	if c.Request.Body != nil {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			var maxBytesErr *http.MaxBytesError
			if errors.As(err, &maxBytesErr) {
				return "", apperrors.ErrPayloadTooLarge.Wrap(err)
			}
			return "", apperrors.ErrMalformedBody.Wrap(err)
		}

		if len(bytes.TrimSpace(body)) > 0 {
			// BindJSON reads the body from this key, since it has already been read here.
			c.Set(gin.BodyBytesKey, body)

			req, err := validator.BindJSON[dto.RefreshRequest](c)
			if err != nil {
				return "", err
			}

			if req.RefreshToken != "" {
				return req.RefreshToken, nil
			}
		}
	}

	cookie, err := c.Cookie(RefreshCookieName)
	if err != nil {
		return "", nil
	}

	return cookie, nil
}
