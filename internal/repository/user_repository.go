package repository

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"

	"github.com/ash2006xo/nullfeed_weblog/internal/model"
)

var ErrUserNotFound = errors.New("user not found")
var ErrUsernameTaken = errors.New("username already taken")

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(username, passwordHash string) (*model.User, error) {
	query := `
		INSERT INTO users (username, password_hash)
		VALUES ($1, $2)
		RETURNING id, username, password_hash, created_at
	`

	var u model.User
	err := r.db.QueryRow(query, username, passwordHash).Scan(
		&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil, ErrUsernameTaken
		}
		return nil, err
	}

	return &u, nil
}

func (r *UserRepository) GetByUsername(username string) (*model.User, error) {
	query := `
		SELECT id, username, password_hash, created_at
		FROM users
		WHERE username = $1
	`

	var u model.User
	err := r.db.QueryRow(query, username).Scan(
		&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &u, nil
}
