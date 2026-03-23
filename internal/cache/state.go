package cache

import "context"

type MatchStateCache interface {
	SetLatestMove(ctx context.Context, matchID int32, move []byte) error
}
