package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// para criar poool de conexão com o banco
func Conectar(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("ERRO ao criar o pool de conexoes: %w", err)
	}

	// banco responde?
	err = pool.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("ERRO ao conectar com banco: %w", err)
	}

	return pool, nil
}
