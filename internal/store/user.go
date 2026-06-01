package store

import (
	"context"
	"database/sql"
	"errors"
)

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserStore struct {
	DB *sql.DB
}

func (s *UserStore) Create(ctx context.Context, username string, password string) (int, error) {
	var userID int
	err := s.DB.QueryRowContext(ctx, `INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id`, username, password).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("user not found")
		}
		return 0, err
	}
	return userID, nil
}

func (s *UserStore) GetByUsername(ctx context.Context, username string) (*User, error) {
	var user User
	err := s.DB.QueryRowContext(ctx, `SELECT id, username, password FROM users WHERE username = $1`, username).Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}
