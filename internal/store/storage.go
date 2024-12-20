package store

import (
	"context"
	"database/sql"
)

type Storage struct {
	Posts interface {
		CreatePost(context.Context, *Post) error
	}
	Users interface {
		CreateUser(context.Context, *User) error
	}
}

func NewStore(db *sql.DB) Storage {
	return Storage{
		Posts: &PostStore{db},
		Users: &UserStore{db},
	}
}
