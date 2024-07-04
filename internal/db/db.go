package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/maxpain/shortener/config"
)

type Database struct {
	*pgxpool.Pool
}

func NewDatabase(cfg *config.Configuration) (*Database, error) {
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseDSN)

	if err != nil {
		return nil, err
	}

	return &Database{pool}, nil
}
