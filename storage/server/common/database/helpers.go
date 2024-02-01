package database

import (
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func IsNotFoundError(err error) bool {
	if errors.Is(err, pgx.ErrNoRows) {
		return true
	}
	return false
}

func IsConflictError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.Is(err, pgErr) {
		if pgErr.Code == "23505" {
			return true
		}
		return false
	}
	return false
}
