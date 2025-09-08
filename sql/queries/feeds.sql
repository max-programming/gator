-- name: CreateFeed :one
INSERT INTO feeds (id, name, url, user_id, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetFeedByURL :one
SELECT *
FROM feeds
WHERE url = $1;

-- name: GetFeeds :many
SELECT f.name,
  f.url,
  u.name AS username
FROM feeds f
  JOIN users u ON u.id = f.user_id;

-- name: MarkFeedFetched :exec
UPDATE feeds
SET last_fetched_at = $1,
  updated_at = $2
WHERE id = $3
  AND user_id = $4;

-- name: GetNextFeedToFetch :one
SELECT *
FROM feeds
WHERE user_id = $1
ORDER BY last_fetched_at ASC NULLS FIRST
LIMIT 1;