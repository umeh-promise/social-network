package store

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
)

type FollowerStore struct {
	db *sql.DB
}

type Follower struct {
	UserID     int64 `json:"user_id"`
	FollowerID int64 `json:"follower_id"`
	CreatedAt  int64 `json:"created_at"`
}

func (store *FollowerStore) Follow(ctx context.Context, followerID, userID int64) error {
	query := `
		INSERT INTO followers (user_id, follower_id) VALUES($1, $2)
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := store.db.ExecContext(ctx, query, userID, followerID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "foreign_key_violation", "unique_violation":
				return ErrorConflict
			default:
				return err
			}
		}
	}
	return nil
}

func (store *FollowerStore) Unfollow(ctx context.Context, followerID, userID int64) error {
	query := `
		DELETE FROM followers
		WHERE user_id = $1 AND follower_id = $2
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := store.db.ExecContext(ctx, query, userID, followerID)
	return err
}
