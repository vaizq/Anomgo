package db

import (
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	ErrCodeUniqueViolation = "23505"
	Unknown                = ""
)

func ErrCode(err error) string {
	var pgErr *pgconn.PgError
	if err == nil || !errors.As(err, &pgErr) {
		return Unknown
	}

	return pgErr.Code
}
