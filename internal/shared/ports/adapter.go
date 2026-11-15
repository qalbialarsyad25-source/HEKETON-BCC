package ports

import (
	"main/internal/shared/ports/hasher"
	"main/internal/shared/ports/logger"
	"main/internal/shared/ports/token"
)

type Adapters struct {
	TokenService token.TokenService
	Hasher       hasher.Hasher
	Logger       logger.Logger
}
