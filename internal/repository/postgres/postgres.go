package postgres

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/maxpain/shortener/internal/model"
	"github.com/maxpain/shortener/internal/repository/postgres/queries"
)

type Repository struct {
	logger  *slog.Logger
	db      *pgxpool.Pool
	queries *queries.Queries
}

// Create a new memory repository with optional persistence to the file.
func New(db *pgxpool.Pool, logger *slog.Logger) *Repository {
	return &Repository{
		logger: logger.With(
			slog.String("repository", "postgres"),
		),
		db:      db,
		queries: queries.New(db),
	}
}

func (r *Repository) Init(ctx context.Context) error {
	schemaFile, err := os.Open("schema.sql")
	if err != nil {
		return fmt.Errorf("failed to open schema.sql: %w", err)
	}
	defer schemaFile.Close()

	sqlBytes, err := io.ReadAll(schemaFile)
	if err != nil {
		return fmt.Errorf("failed to read schema.sql: %w", err)
	}

	sql := string(sqlBytes)

	_, err = r.db.Exec(ctx, sql)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func (r *Repository) Get(ctx context.Context, hash string) (*model.StoredLink, error) {
	row, err := r.queries.SelectLink(ctx, hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}

		return nil, fmt.Errorf("failed to select link: %w", err)
	}

	return &model.StoredLink{
		Hash: hash,
		Link: &model.Link{
			OriginalURL:   row.OriginalUrl,
			CorrelationID: row.CorrelationID,
		},
	}, nil
}

func (r *Repository) Save(ctx context.Context, linksToStore []*model.StoredLink) ([]bool, error) {
	results := make([]bool, 0, len(linksToStore))

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}

	defer tx.Rollback(ctx) //nolint:errcheck

	for _, link := range linksToStore {
		rowsAffected, err := r.queries.WithTx(tx).InsertLink(ctx, queries.InsertLinkParams{
			Hash:          link.Hash,
			OriginalUrl:   link.OriginalURL,
			CorrelationID: link.CorrelationID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to insert link: %w", err)
		}

		isExists := rowsAffected == 0
		results = append(results, isExists)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return results, nil
}

func (r *Repository) Ping(ctx context.Context) error {
	err := r.db.Ping(ctx)

	return fmt.Errorf("failed to ping db: %w", err)
}

func (r *Repository) Close() {
	r.db.Close()
}
