-- name: CreatePost :exec
INSERT INTO posts (
    id,
    title,
    url,
    description,
    published_at,
    feed_id,
    created_at,
    updated_at
  )
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: GetPostsForUser :many
SELECT p.*
FROM posts p
  JOIN feed_follows ff ON ff.feed_id = p.feed_id
WHERE ff.user_id = $1
ORDER BY p.published_at DESC
LIMIT $2;