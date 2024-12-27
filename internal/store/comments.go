package store

import (
	"context"
	"database/sql"
)

type Comment struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	PostID    int64  `json:"post_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	User      User   `json:"user"`
}

type CommentStore struct {
	db *sql.DB
}

func (store *CommentStore) GetByPostID(ctx context.Context, postID int64) ([]Comment, error) {

	query := `
		SELECT c.id, c.post_id, c.user_id, c.content, c.created_at, users.id, users.username from comments c
		JOIN users on users.id= c.user_id
		WHERE c.post_id = $1
		ORDER BY c.created_at;
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := store.db.QueryContext(ctx, query, postID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	comments := []Comment{}

	for rows.Next() {
		var c Comment
		c.User = User{}
		err := rows.Scan(&c.ID, &c.PostID, &c.UserID, &c.Content, &c.CreatedAt, &c.User.ID, &c.User.Username)
		if err != nil {
			return nil, err
		}

		comments = append(comments, c)
	}

	return comments, nil

}

func (store *CommentStore) Create(ctx context.Context, comment *Comment) error {

	query := `
		INSERT INTO comments(user_id, post_id, content)
		VALUES($1, $2, $3)
		RETURNING id, created_at;
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := store.db.QueryRowContext(ctx, query, comment.UserID, comment.PostID, comment.Content).Scan(&comment.ID, &comment.CreatedAt)

	if err != nil {
		return err
	}

	return nil
}
