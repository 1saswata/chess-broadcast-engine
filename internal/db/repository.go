package db

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type UserRepository struct {
	d *sql.DB
}

type User struct {
	ID           string
	UserName     string
	PasswordHash string
	Role         string
	CreatedAt    string
}

func (u UserRepository) CreateUser(ctx context.Context, username string,
	passwordHash string, role string) error {
	id := uuid.New()
	_, err := u.d.Exec(`INSERT INTO users (id, username, password_hash, role) 
		VALUES (?, ?, ?, ?)`, id, username, passwordHash, role)
	return err
}

func (u UserRepository) GetUserByUsername(ctx context.Context,
	username string) (User, error) {
	ud := User{}
	row := u.d.QueryRow("SELECT * FROM users WHERE username = ?", username)
	err := row.Scan(&ud.ID, &ud.UserName, &ud.PasswordHash, &ud.Role, &ud.CreatedAt)
	return ud, err
}
