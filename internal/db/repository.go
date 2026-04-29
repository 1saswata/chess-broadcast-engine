package db

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type UserRepository struct {
	D *sql.DB
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
	_, err := u.D.Exec(`INSERT INTO users (id, username, password_hash, role) 
		VALUES ($1, $2, $3, $4)`, id, username, passwordHash, role)
	return err
}

func (u UserRepository) GetUserByUsername(ctx context.Context,
	username string) (User, error) {
	ud := User{}
	row := u.D.QueryRow("SELECT * FROM users WHERE username = $1", username)
	err := row.Scan(&ud.ID, &ud.UserName, &ud.PasswordHash, &ud.Role, &ud.CreatedAt)
	return ud, err
}
