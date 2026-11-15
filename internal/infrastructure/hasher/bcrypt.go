package hasher

import (
	"errors"

	"main/internal/config"
	"main/internal/shared/ports/hasher"

	"golang.org/x/crypto/bcrypt"
)

type BcryptHasher struct {
	cost int
}

func NewBcryptHasher(cfg *config.BcryptConfig) hasher.Hasher {
	return &BcryptHasher{cost: cfg.Cost}
}

func (b BcryptHasher) Compare(hashed string, literal string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(literal))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return hasher.ErrHashMismatch
	}

	return err
}

func (b BcryptHasher) Hash(value string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(value), b.cost)
	if err != nil {
		return "", err
	}

	return string(hashed), nil
}

var _ hasher.Hasher = (*BcryptHasher)(nil)
