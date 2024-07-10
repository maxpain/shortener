-- name: SelectLink :one
SELECT *
FROM links
WHERE hash = $1;

-- name: SelectUserLinks :many
SELECT *
FROM links
WHERE user_id = $1;

-- name: InsertLink :execrows
INSERT INTO links (hash, original_url, correlation_id, user_id)
VALUES ($1, $2, $3, $4)
ON CONFLICT (hash) DO NOTHING;

-- name: MarkLinksAsDeleted :exec
UPDATE links
SET is_deleted = true
WHERE user_id = $1 AND hash = ANY(sqlc.arg('hashes')::text[]);