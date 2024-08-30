package postgresql

import (
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func IsConflictErr(err error) bool {
	if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
		return true
	}
	return false
}

var ErrNoRows = pgx.ErrNoRows
