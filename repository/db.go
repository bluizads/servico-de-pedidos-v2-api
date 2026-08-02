package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DB interface {
	Begin(contexto context.Context) (pgx.Tx, error)
	Exec(contexto context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(contexto context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(contexto context.Context, sql string, args ...any) pgx.Row
}
