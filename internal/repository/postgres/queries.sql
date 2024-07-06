-- name: SelectLink :one
SELECT
	original_url,
	correlation_id
FROM links
WHERE hash = $1;

-- name: InsertLink :execrows
INSERT INTO links (hash, original_url, correlation_id)
VALUES ($1, $2, $3)
ON CONFLICT (hash) DO NOTHING;