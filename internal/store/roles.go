package store

import (
	"context"
	"database/sql"
)

type Role struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Level       int64  `json:"level"`
}

type RoleStore struct {
	db *sql.DB
}

func (store *RoleStore) GetByName(ctx context.Context, roleName string) (*Role, error) {
	query := `
	SELECT id, name, description, level FROM roles
	WHERE name = $1
	`
	role := &Role{}

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := store.db.QueryRowContext(ctx, query, roleName).Scan(
		&role.ID,
		&role.Name,
		&role.Description,
		&role.Level,
	)

	if err != nil {
		return nil, err
	}

	return role, nil
}
