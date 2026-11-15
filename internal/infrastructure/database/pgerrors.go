package database

import (
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

const (
	UniqueViolation     = "23505"
	ForeignKeyViolation = "23503"
)

func ConstraintViolation(err error) (code string, constraint string, ok bool) {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || !strings.HasPrefix(pgErr.Code, "23") {
		return "", "", false
	}

	return pgErr.Code, pgErr.ConstraintName, true
}
