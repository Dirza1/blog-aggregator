-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at,name, url,user_id)
VALUES(
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;

-- name: GetFeeds :many
SELECT feeds.name,url,users.name AS username
FROM feeds
LEFT JOIN users
ON feeds.user_id = users.id;

-- name: GetFeedId :one
SELECT id FROM feeds
WHERE url = $1;