package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/1saswata/chess-broadcast-engine/internal/pb"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"google.golang.org/protobuf/proto"
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
	if err != nil {
		if err == sql.ErrNoRows {
			return ud, fmt.Errorf("user not found")
		}
		return ud, err
	}
	return ud, nil
}

func (u UserRepository) ArchiveMatch(ctx context.Context, matchID string,
	whitePlayerID, blackPlayerID string, moveData [][]byte) error {
	tx, err := u.D.BeginTx(ctx, nil)
	_, err = tx.Exec(`INSERT INTO matches (id, white_player_id, black_player_id, 
	status) VALUES ($1, $2, $3, "completed")`, matchID, whitePlayerID,
		blackPlayerID)
	if err != nil {
		tx.Rollback()
		return err
	}
	for _, move := range moveData {
		id := uuid.New()
		var m pb.Move
		err := proto.Unmarshal(move, &m)
		if err != nil {
			tx.Rollback()
			return err
		}
		_, err = tx.Exec(`INSERT INTO moves (id, match_id, sequence_number, 
		move_payload) VALUES ($1, $2, $3, $4)`, id, matchID, m.SequenceNumber,
			string(move))
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
