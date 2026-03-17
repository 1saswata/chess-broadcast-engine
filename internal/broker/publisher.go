package broker

import (
	"context"
)

type MovePublisher interface {
	PublishMove(ctx context.Context, move []byte) error
}
