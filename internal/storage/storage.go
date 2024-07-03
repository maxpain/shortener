package storage

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/maxpain/shortener/config"
	"github.com/maxpain/shortener/internal/db"
	"github.com/maxpain/shortener/internal/log"
	"github.com/maxpain/shortener/internal/utils"
	"go.uber.org/zap"

	"github.com/jackc/pgx/v5"
)

const length = 7

type LinkMap map[string]Link

type Storage struct {
	logger  *log.Logger
	links   LinkMap
	file    *os.File
	DB      *db.Database
	baseURL string
	mu      sync.Mutex
}

type Link struct {
	Hash          string `json:"hash"`
	OriginalURL   string `json:"original_url"`
	CorrelationID string `json:"correlation_id"`
}

func NewStorage(cfg *config.Configuration, logger *log.Logger) (*Storage, error) {
	s := &Storage{
		logger:  logger,
		links:   make(LinkMap),
		file:    nil,
		baseURL: cfg.BaseURL,
	}

	if cfg.DatabaseDSN != "" {
		var err error
		s.DB, err = db.NewDatabase(cfg)

		if err != nil {
			return nil, err
		}

		_, err = s.DB.Exec(context.Background(), `
			CREATE TABLE IF NOT EXISTS links (
				hash TEXT PRIMARY KEY,
				original_url TEXT NOT NULL,
				correlation_id TEXT
			);

			CREATE INDEX IF NOT EXISTS correlation_id_idx ON links (correlation_id);
		`)

		if err != nil {
			return nil, err
		}
	} else if cfg.FileStoragePath != "" {
		file, err := os.OpenFile(cfg.FileStoragePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)

		if err != nil {
			return nil, err
		}

		s.file = file

		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			line := scanner.Text()

			var link Link
			err := json.Unmarshal([]byte(line), &link)

			if err != nil {
				return nil, err
			}

			logger.Debug("Loaded link",
				zap.String("hash", link.Hash),
				zap.String("original_url", link.OriginalURL),
				zap.String("correlation_id", link.CorrelationID),
			)

			s.links[link.Hash] = link
		}
	}

	return s, nil
}

func (s *Storage) Close() error {
	if s.file != nil {
		return s.file.Close()
	}

	if s.DB != nil {
		s.DB.Close()
	}

	return nil
}

type LinkInput struct {
	OriginalURL   string
	CorrelationID string
}

type LinkOutput struct {
	ShortURL      string
	CorrelationID string
	AlreadyExists bool
}

func (s *Storage) Save(
	ctx context.Context,
	links []LinkInput,
) ([]LinkOutput, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	shortenedLinks := make([]LinkOutput, 0, len(links))

	var tx pgx.Tx

	if s.DB != nil {
		var err error
		tx, err = s.DB.Begin(ctx)

		if err != nil {
			return nil, err
		}

		defer tx.Rollback(ctx)
	}

	for _, link := range links {
		s.logger.Debug("Saving link",
			zap.String("url", link.OriginalURL),
			zap.String("correlation_id", link.CorrelationID),
		)

		shortenedLink := LinkOutput{
			CorrelationID: link.CorrelationID,
		}

		hash := generateHash(link.OriginalURL)

		if s.DB != nil {
			ct, err := tx.Exec(ctx, `
				INSERT INTO links (hash, original_url, correlation_id)
				VALUES ($1, $2, $3)
				ON CONFLICT (hash) DO NOTHING
			`, hash, link.OriginalURL, link.CorrelationID)

			if err != nil {
				s.logger.Error("Failed to insert link to database",
					zap.Error(err),
					zap.String("url", link.OriginalURL),
					zap.String("correlation_id", link.CorrelationID),
				)

				return nil, err
			}

			if ct.RowsAffected() == 0 {
				shortenedLink.AlreadyExists = true
			}
		} else {
			_, ok := s.links[hash]

			if ok {
				shortenedLink.AlreadyExists = true
			} else {
				linkToSave := Link{
					OriginalURL:   link.OriginalURL,
					Hash:          hash,
					CorrelationID: link.CorrelationID,
				}

				if s.file != nil {
					jsonString, err := json.Marshal(linkToSave)

					if err != nil {
						return nil, err
					}

					_, err = s.file.WriteString(string(jsonString) + "\n")

					if err != nil {
						return nil, err
					}
				}

				s.links[hash] = linkToSave
			}
		}

		var err error
		shortenedLink.ShortURL, err = utils.СonstructURL(s.baseURL, hash)

		if err != nil {
			return nil, errors.New("failed to construct URL")
		}

		shortenedLinks = append(shortenedLinks, shortenedLink)
	}

	if s.DB != nil {
		err := tx.Commit(ctx)

		if err != nil {
			return nil, err
		}
	}

	return shortenedLinks, nil
}

func (s *Storage) GetURL(ctx context.Context, hash string) (string, error) {
	if s.DB == nil {
		link, ok := s.links[hash]

		if !ok {
			return "", nil
		}

		return link.OriginalURL, nil
	} else {
		var url string
		err := s.DB.QueryRow(ctx, "SELECT original_url FROM links WHERE hash = $1", hash).Scan(&url)

		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return "", nil
			}

			return "", err
		}

		return url, nil
	}
}

func generateHash(url string) string {
	hash := sha256.Sum256([]byte(url))
	return hex.EncodeToString(hash[:])[:length]
}
