package dto

type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=1,max=100"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

// LoginRequest checks only that the fields are there: a wrong password is INVALID_CREDENTIALS,
// whatever its length.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RefreshRequest is the optional body of refresh and logout, for clients without a cookie jar
// (mobile apps, scripts). When it is missing or empty, the refresh_token cookie is used.
type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}
