-- name: CreatePost :exec
INSERT INTO posts (id, created_at, updated_at, title,url,description,published_at,feed_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
);

-- name: GetPostsFor User :$1
SELECT * FROM posts
WHERE 