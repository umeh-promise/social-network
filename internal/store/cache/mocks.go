package cache

import (
	"context"

	"github.com/umeh-promise/social/internal/store"
)

func NewMockStore() Storage {
	return Storage{
		Users: &MockUserStore{},
	}
}

type MockUserStore struct{}

func (m *MockUserStore) Get(context.Context, int64) (*store.User, error) {
	return nil, nil
}

func (m *MockUserStore) Set(context.Context, *store.User) error {
	return nil
}
