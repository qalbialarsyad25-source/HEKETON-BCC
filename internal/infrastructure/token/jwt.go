package token

import (
	"errors"
	"fmt"
	"main/internal/config"
	apperrors "main/internal/shared/errors"
	"main/internal/shared/ports/token"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTService struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

// Token types, stored in the "typ" claim. In DEV both tokens may share one secret,
// so the type is what keeps a refresh token from being used as an access token.
const (
	TypeAccess  = "access"
	TypeRefresh = "refresh"
)

const issuer = "go-starter"

type Claims struct {
	Payload   token.JWTPayload
	TokenType string `json:"typ"`

	jwt.RegisteredClaims
}

func NewTokenService(cfg *config.JWTConfig) token.TokenService {
	return &JWTService{
		accessSecret:  []byte(cfg.AccessSecret),
		refreshSecret: []byte(cfg.RefreshSecret),
		accessTTL:     time.Duration(cfg.AccessExpiryMinutes) * time.Minute,
		refreshTTL:    time.Duration(cfg.RefreshExpiryDays) * 24 * time.Hour,
	}
}

func (js *JWTService) AccessTTL() time.Duration {
	return js.accessTTL
}

func (js *JWTService) RefreshTTL() time.Duration {
	return js.refreshTTL
}

func (js *JWTService) GenerateAccessToken(data *token.JWTPayload) (*token.JWT, error) {
	return generate(data, TypeAccess, js.accessSecret, js.accessTTL)
}

func (js *JWTService) GenerateRefreshToken(data *token.JWTPayload) (*token.JWT, error) {
	return generate(data, TypeRefresh, js.refreshSecret, js.refreshTTL)
}

// generate signs a token of the given type with a fresh JTI. Subject is the user id and ID the JTI,
// so standard JWT tooling can read them too.
func generate(data *token.JWTPayload, tokenType string, secret []byte, ttl time.Duration) (*token.JWT, error) {
	payload := token.JWTPayload{
		JTI:    uuid.New(),
		UserID: data.UserID,
		Name:   data.Name,
	}

	now := time.Now()
	// JWT times have second precision, so ExpiresAt is truncated to match what Parse checks.
	expiresAt := now.Add(ttl).Truncate(time.Second)

	claims := Claims{
		Payload:   payload,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   payload.UserID.String(),
			ID:        payload.JTI.String(),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    issuer,
		},
	}

	signed := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := signed.SignedString(secret)

	if err != nil {
		return nil, err
	}

	token := token.JWT{Payload: payload, Value: tokenString, ExpiresAt: expiresAt}

	return &token, nil
}

func (js *JWTService) ParseAccessToken(tokenString string) (*token.JWTPayload, error) {
	return parse(tokenString, js.accessSecret, TypeAccess)
}

func (js *JWTService) ParseRefreshToken(tokenString string) (*token.JWTPayload, error) {
	return parse(tokenString, js.refreshSecret, TypeRefresh)
}

// parse verifies an HS256 token of tokenType signed with secret. An expired token is ErrTokenExpired;
// every other failure (bad signature, other algorithm, other token type, garbage) is ErrInvalidToken.
// Both are 401s and keep the jwt error as the cause for the log.
func parse(tokenString string, secret []byte, tokenType string) (*token.JWTPayload, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(t *jwt.Token) (interface{}, error) {
			// Only HS256 is accepted, so a token can't pick its own algorithm (alg=none, HS512, RS256).
			if t.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
			}
			return secret, nil
		},
	)

	if errors.Is(err, jwt.ErrTokenExpired) {
		return nil, apperrors.ErrTokenExpired.Wrap(err)
	}

	if err != nil {
		return nil, apperrors.ErrInvalidToken.Wrap(err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, apperrors.ErrInvalidToken.Wrap(errors.New("unexpected claims type"))
	}

	if claims.TokenType != tokenType {
		return nil, apperrors.ErrInvalidToken.Wrap(fmt.Errorf("token type %q, want %q", claims.TokenType, tokenType))
	}

	payload := claims.Payload

	return &payload, nil
}

var _ token.TokenService = (*JWTService)(nil)
