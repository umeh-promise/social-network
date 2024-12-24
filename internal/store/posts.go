package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

type Post struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	UserID    int64     `json:"user_id"`
	Tags      []string  `json:"tags"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	Version   int       `json:"version"`
	Comments  []Comment `json:"comments"`
}

type PostStore struct {
	db *sql.DB
}

func (store *PostStore) Create(ctx context.Context, post *Post) error {
	query := `
		INSERT INTO posts (title, content, user_id, tags)
		VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at
	`

	err := store.db.QueryRowContext(ctx, query,
		post.Title,
		post.Content,
		post.UserID,
		pq.Array(post.Tags),
	).Scan(
		&post.ID,
		&post.CreatedAt,
		&post.UpdatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

func (store *PostStore) GetByID(ctx context.Context, id int64) (*Post, error) {
	var post Post

	query := `
		SELECT id, user_id, title, content, tags, created_at, updated_at, version FROM posts
		WHERE id=$1
	`

	err := store.db.QueryRowContext(ctx, query, id).Scan(
		&post.ID,
		&post.UserID,
		&post.Title,
		&post.Content,
		pq.Array(&post.Tags),
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrorNotFound

		default:
			return nil, err
		}
	}

	return &post, nil
}

func (store *PostStore) Update(ctx context.Context, post *Post) error {
	query := `
		UPDATE posts
		SET title = $1, content = $2, tags = $3, version = version + 1
		WHERE id = $4 AND version = $5
		RETURNING version;
	`

	err := store.db.QueryRowContext(ctx, query, post.Title, post.Content, pq.Array(post.Tags), post.ID, post.Version).Scan(&post.Version)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrorNotFound

		default:
			return err
		}

	}

	return nil
}

func (store *PostStore) Delete(ctx context.Context, id int64) error {

	query := `
		DELETE FROM posts
		WHERE id = $1;
	`

	res, err := store.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrorNotFound
	}

	return nil
}
