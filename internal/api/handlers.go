package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/1saswata/chess-broadcast-engine/internal/auth"
	"github.com/google/uuid"
)

func (s *APIServer) HandleRegister(w http.ResponseWriter, r *http.Request) {
	v := struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	hPass, err := auth.HashPassword(v.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	err = s.Repo.CreateUser(ctx, v.Username, hPass, v.Role)
	if err != nil {
		if err == context.DeadlineExceeded {
			http.Error(w, "Database is slow", http.StatusGatewayTimeout)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (s *APIServer) HandleLogin(w http.ResponseWriter, r *http.Request) {
	v := struct {
		Username string `json:"username"`
		Password string `json:"password"`
		MatchID  int32  `json:"match_id"`
	}{}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	err := json.NewDecoder(r.Body).Decode(&v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	user, err := s.Repo.GetUserByUsername(ctx, v.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	if !auth.CheckPasswordHash(v.Password, user.PasswordHash) {
		http.Error(w, "Wrong password", http.StatusUnauthorized)
		return
	}
	token, err := auth.GenerateToken(user.ID, v.MatchID, user.Role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "%s", token)
}

func (s *APIServer) HandleArchive(w http.ResponseWriter, r *http.Request) {
	v := struct {
		MatchID       int32     `json:"match_id"`
		WhitePlayerId uuid.UUID `json:"white_player_id"`
		BlackPlayerID uuid.UUID `json:"black_player_id"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	moveHistory, err := s.Cache.GetMoveHistory(ctx, v.MatchID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = s.Repo.ArchiveMatch(ctx, v.MatchID, v.WhitePlayerId, v.BlackPlayerID,
		moveHistory)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	key := fmt.Sprintf("match:%d:latest", v.MatchID)
	err = s.Cache.DeleteKey(ctx, key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	key = fmt.Sprintf("match:%d:sequence", v.MatchID)
	err = s.Cache.DeleteKey(ctx, key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *APIServer) HandleProvisionMatch(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "missing auth header", http.StatusUnauthorized)
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := auth.ValidateToken(token)
	if err != nil || claims["role"] != "admin" {
		http.Error(w, "admin access required", http.StatusForbidden)
		return
	}
	v := struct {
		MatchID       int32     `json:"match_id"`
		WhitePlayerId uuid.UUID `json:"white_player_id"`
		BlackPlayerID uuid.UUID `json:"black_player_id"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := s.Repo.ProvisionMatch(ctx, v.MatchID, v.WhitePlayerId,
		v.BlackPlayerID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := s.Cache.AuthorizePlayers(ctx, v.MatchID, v.WhitePlayerId.String(),
		v.BlackPlayerID.String()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
