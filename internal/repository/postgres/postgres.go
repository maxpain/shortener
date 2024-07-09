package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/maxpain/shortener/internal/model"
	"github.com/maxpain/shortener/internal/repository/postgres/queries"
)

type Repository struct {
	logger   *slog.Logger
	db       *pgxpool.Pool
	queries  *queries.Queries
	deleteCh chan DeletionRequest
}

type DeletionRequest struct {
	Hashes []string
	UserID string
}

// Create a new memory repository with optional persistence to the file.
func New(db *pgxpool.Pool, logger *slog.Logger) *Repository {
	return &Repository{
		logger: logger.With(
			slog.String("repository", "postgres"),
		),
		db:       db,
		queries:  queries.New(db),
		deleteCh: make(chan DeletionRequest, 1024),
	}
}

func (r *Repository) Init(ctx context.Context) error {
	_, err := r.db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS links (
			hash VARCHAR(6) PRIMARY KEY,
			original_url TEXT NOT NULL,
			correlation_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			is_deleted BOOLEAN DEFAULT FALSE NOT NULL
		);

		CREATE INDEX IF NOT EXISTS correlation_id_idx ON links (correlation_id);
		CREATE INDEX IF NOT EXISTS user_id_idx ON links (user_id);
	`)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	go r.deleteLoop()

	return nil
}

func (r *Repository) GetLink(ctx context.Context, hash string) (*model.StoredLink, error) {
	row, err := r.queries.SelectLink(ctx, hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}

		return nil, fmt.Errorf("failed to select link: %w", err)
	}

	return &model.StoredLink{
		Hash:      row.Hash,
		UserID:    row.UserID,
		IsDeleted: row.IsDeleted,
		Link: &model.Link{
			OriginalURL:   row.OriginalUrl,
			CorrelationID: row.CorrelationID,
		},
	}, nil
}

func (r *Repository) GetUserLinks(ctx context.Context, userID string) ([]*model.StoredLink, error) {
	rows, err := r.queries.SelectUserLinks(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to select link: %w", err)
	}

	links := make([]*model.StoredLink, 0, len(rows))

	for _, row := range rows {
		links = append(links, &model.StoredLink{
			Hash:      row.Hash,
			UserID:    row.UserID,
			IsDeleted: row.IsDeleted,
			Link: &model.Link{
				OriginalURL:   row.OriginalUrl,
				CorrelationID: row.CorrelationID,
			},
		})
	}

	return links, nil
}

func (r *Repository) SaveLinks(ctx context.Context, linksToStore []*model.StoredLink) ([]bool, error) {
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
			UserID:        link.UserID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to insert link: %w", err)
		}

		isExists := rowsAffected == 0
		results = append(results, !isExists)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return results, nil
}

func (r *Repository) MarkForDeletion(hashes []string, userID string) error {
	r.deleteCh <- DeletionRequest{
		Hashes: hashes,
		UserID: userID,
	}

	return nil
}

func (r *Repository) deleteLoop() {
	for req := range r.deleteCh {
		err := r.queries.MarkLinksAsDeleted(context.Background(), queries.MarkLinksAsDeletedParams{
			Hashes: req.Hashes,
			UserID: req.UserID,
		})
		if err != nil {
			r.logger.Error("failed to mark links as deleted", slog.Any("error", err))
		}
	}
}

func (r *Repository) Ping(ctx context.Context) error {
	err := r.db.Ping(ctx)
	if err != nil {
		return fmt.Errorf("failed to ping db: %w", err)
	}

	return nil
}

func (r *Repository) Close() error {
	r.db.Close()
	close(r.deleteCh)

	return nil
}
