package dto

import (
	"main/internal/modules/auth/domain"
	userdto "main/internal/modules/user/dto"
)

// SessionResponse is the body of register, login and refresh. The refresh token is also set
// as an httpOnly cookie; browsers should rely on the cookie and ignore it here.
type SessionResponse struct {
	AccessToken  string               `json:"accessToken"`
	RefreshToken string               `json:"refreshToken"`
	TokenType    string               `json:"tokenType"`
	ExpiresIn    int                  `json:"expiresIn"` // seconds until the access token expires
	User         userdto.UserResponse `json:"user"`
}

func SessionToResponse(s domain.Session) *SessionResponse {
	return &SessionResponse{
		AccessToken:  s.AccessToken,
		RefreshToken: s.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(s.AccessExpiresIn.Seconds()),
		User:         *userdto.DomainToResponse(s.User),
	}
}
