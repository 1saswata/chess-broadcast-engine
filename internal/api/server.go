package api

import (
	"github.com/1saswata/chess-broadcast-engine/internal/cache"
	"github.com/1saswata/chess-broadcast-engine/internal/db"
)

type APIServer struct {
	Repo  db.UserRepository
	Cache *cache.RedisCache
}

func NewAPIServer(repo db.UserRepository, cache *cache.RedisCache) *APIServer {
	return &APIServer{Repo: repo, Cache: cache}
}
