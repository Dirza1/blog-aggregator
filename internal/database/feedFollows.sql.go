// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: feedFollows.sql

package database

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const createFeedFollow = `-- name: CreateFeedFollow :one
with inserted_feed_follow AS (
    INSERT INTO feed_follows (id, created_at,updated_at,user_id,feed_id)
    VALUES(
    $1,
    $2,
    $3,
    $4,
    $5
    )
    RETURNING id, created_at, updated_at, user_id, feed_id
)
SELECT
    inserted_feed_follow.id, inserted_feed_follow.created_at, inserted_feed_follow.updated_at, inserted_feed_follow.user_id, inserted_feed_follow.feed_id,
    feeds.name AS feed_name,
    users.name AS user_name
    FROM inserted_feed_follow
    INNER JOIN users
    ON users.id = inserted_feed_follow.user_id
    INNER JOIN feeds
    ON feeds.id = inserted_feed_follow.feed_id
`

type CreateFeedFollowParams struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	UserID    uuid.UUID
	FeedID    uuid.UUID
}

type CreateFeedFollowRow struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	UserID    uuid.UUID
	FeedID    uuid.UUID
	FeedName  string
	UserName  string
}

func (q *Queries) CreateFeedFollow(ctx context.Context, arg CreateFeedFollowParams) (CreateFeedFollowRow, error) {
	row := q.db.QueryRowContext(ctx, createFeedFollow,
		arg.ID,
		arg.CreatedAt,
		arg.UpdatedAt,
		arg.UserID,
		arg.FeedID,
	)
	var i CreateFeedFollowRow
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.UserID,
		&i.FeedID,
		&i.FeedName,
		&i.UserName,
	)
	return i, err
}

const getFeedFollowsForUser = `-- name: GetFeedFollowsForUser :many
SELECT
    users.name AS user_name,
    feeds.name AS feed_name
    FROM feed_follows
    INNER JOIN users
    ON users.id = feed_follows.user_id
    INNER JOIN feeds
    ON feeds.id = feed_follows.feed_id
    WHERE feed_follows.user_id = $1
`

type GetFeedFollowsForUserRow struct {
	UserName string
	FeedName string
}

func (q *Queries) GetFeedFollowsForUser(ctx context.Context, userID uuid.UUID) ([]GetFeedFollowsForUserRow, error) {
	rows, err := q.db.QueryContext(ctx, getFeedFollowsForUser, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetFeedFollowsForUserRow
	for rows.Next() {
		var i GetFeedFollowsForUserRow
		if err := rows.Scan(&i.UserName, &i.FeedName); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const removeFollow = `-- name: RemoveFollow :exec
DELETE 
    FROM feed_follows
    WHERE user_id IN (
        SELECT id
        FROM users
        WHERE users.name = $1)
        AND feed_id IN (
            SELECT id
            FROM feeds
            WHERE url = $2
        )
`

type RemoveFollowParams struct {
	Name string
	Url  string
}

func (q *Queries) RemoveFollow(ctx context.Context, arg RemoveFollowParams) error {
	_, err := q.db.ExecContext(ctx, removeFollow, arg.Name, arg.Url)
	return err
}
