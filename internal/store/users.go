package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrorDuplicateEmail    = errors.New("a user with that email already exists")
	ErrorDuplicateUsername = errors.New("a user with that username already exists")
)

type User struct {
	ID        int64    `json:"id"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Password  password `json:"-"`
	CreatedAt string   `json:"created_at"`
	IsActive  bool     `json:"is_active"`
	RoleID    int64    `json:"role_id"`
	Role      Role     `json:"role"`
}

type password struct {
	text *string
	hash []byte
}

func (p *password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	p.text = &text
	p.hash = hash

	return nil
}

type UserStore struct {
	db *sql.DB
}

func (store *UserStore) Create(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `
		INSERT INTO users (username, email, password,role_id)
		VALUES ($1, $2, $3, (SELECT id FROM roles WHERE name = $4)) RETURNING id, created_at
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	role := user.Role.Name
	if role == "" {
		role = "user"
	}

	err := tx.QueryRowContext(ctx, query,
		user.Username,
		user.Email,
		user.Password.hash,
		role,
	).Scan(
		&user.ID,
		&user.CreatedAt,
	)

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrorDuplicateEmail
		case err.Error() == `pq: duplicate key value violates unique constraint "users_username_key"`:
			return ErrorDuplicateUsername
		default:
			return err
		}
	}

	return nil
}

func (store *UserStore) GetByID(ctx context.Context, userId int64) (*User, error) {
	var user User
	query := `
		SELECT users.id, username, email, created_at, roles.* FROM users
		JOIN roles ON (users.role_id = roles.id)
		WHERE users.id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := store.db.QueryRowContext(ctx, query, userId).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt,
		&user.Role.ID,
		&user.Role.Name,
		&user.Role.Level,
		&user.Role.Description,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrorNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (store *UserStore) CreateAndInvite(ctx context.Context, user *User, token string, invitationExp time.Duration) error {
	return WithTx(store.db, ctx, func(tx *sql.Tx) error {
		// create user
		if err := store.Create(ctx, tx, user); err != nil {
			return err
		}

		// create user invitation
		if err := store.createUserInvitation(ctx, tx, token, invitationExp, user.ID); err != nil {
			return err
		}

		return nil
	})
}

func (store *UserStore) createUserInvitation(ctx context.Context, tx *sql.Tx, token string, exp time.Duration, userID int64) error {

	query := `INSERT INTO user_invitations(token, user_id, expiry) VALUES($1, $2, $3)`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, token, userID, time.Now().Add(exp))
	if err != nil {
		return err
	}

	return nil
}

func (store *UserStore) Activate(ctx context.Context, token string) error {

	return WithTx(store.db, ctx, func(tx *sql.Tx) error {
		// find the user that this token belongs to
		user, err := store.getUserFromInvitation(ctx, tx, token)
		if err != nil {
			return err
		}

		// update th user
		user.IsActive = true
		if err := store.update(ctx, tx, user); err != nil {
			return err
		}

		// delete the invitations
		if err := store.deleteUserInvitation(ctx, tx, user.ID); err != nil {
			return err
		}
		return nil
	})
}

func (store *UserStore) getUserFromInvitation(ctx context.Context, tx *sql.Tx, token string) (*User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.created_at, u.is_active FROM users u
		JOIN user_invitations ui ON u.id = ui.user_id
		WHERE ui.token = $1 AND ui.expiry > $2
	`

	hash := sha256.Sum256([]byte(token))
	hashToken := hex.EncodeToString(hash[:])

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	user := &User{}
	err := tx.QueryRowContext(ctx, query, hashToken, time.Now()).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
		&user.IsActive,
	)

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrorNotFound
		default:
			return nil, err
		}
	}

	return user, nil
}

func (store *UserStore) update(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `UPDATE users  SET username = $1, email = $2, is_active = $3 WHERE id = $4`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, user.Username, user.Email, user.IsActive, user.ID)
	if err != nil {
		return err
	}

	return nil
}

func (store *UserStore) deleteUserInvitation(ctx context.Context, tx *sql.Tx, userID int64) error {
	query := `DELETE FROM user_invitations WHERE user_id = $1`

	_, err := tx.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}

	return nil
}

func (store *UserStore) Delete(ctx context.Context, userID int64) error {
	return WithTx(store.db, ctx, func(tx *sql.Tx) error {
		if err := store.delete(ctx, tx, userID); err != nil {
			return err
		}

		if err := store.deleteUserInvitation(ctx, tx, userID); err != nil {
			return err
		}

		return nil
	})
}

func (store *UserStore) delete(ctx context.Context, tx *sql.Tx, id int64) error {
	query := `DELETE FROM users where id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}

func (store *UserStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	query := `
		SELECT id, email, username, created_at FROM users WHERE email = $1;
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := store.db.QueryRowContext(ctx, query, email).Scan(&user.ID, &user.Email, &user.Username, &user.CreatedAt)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrorNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}
