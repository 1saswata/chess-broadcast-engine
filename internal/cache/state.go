package cache

import "context"

type MatchStateCache interface {
	AppendMove(ctx context.Context, matchID int32, move []byte) error
	GetMoveHistory(ctx context.Context, matchID int32) ([][]byte, error)
}
