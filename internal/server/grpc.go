package server

import (
	"context"
	"log/slog"

	"github.com/1saswata/chess-broadcast-engine/internal/broker"
	"github.com/1saswata/chess-broadcast-engine/internal/cache"
	"github.com/1saswata/chess-broadcast-engine/internal/pb"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type IngestServer struct {
	pb.UnimplementedChessIngestServiceServer
	publisher broker.MovePublisher
	mCache    cache.MatchStateCache
}

func (is *IngestServer) RecordMove(ctx context.Context, m *pb.Move) (*pb.MoveResponse, error) {
	ctx, span := otel.Tracer("ingest-node").Start(ctx, "RecordMove")
	defer span.End()
	b, err := proto.Marshal(m)
	if err != nil {
		slog.Error("Unable to serialize", "Error", err)
		return &pb.MoveResponse{Success: false, Msg: "Internal Server Error"},
			status.Errorf(codes.Internal, "failed to serialize: %v", err)
	}
	span.SetAttributes(attribute.Int("match.id", int(m.GetMatchId())))
	if err = is.mCache.AppendMove(ctx, m.GetMatchId(), b); err != nil {
		slog.Error("Unable to cache", "Error", err)
		return &pb.MoveResponse{Success: false, Msg: "Internal Server Error"},
			status.Errorf(codes.Internal, "failed to cache: %v", err)
	}
	if err = is.publisher.PublishMove(ctx, b); err != nil {
		slog.Error("Unable to publish but move cached", "Error", err)
		return &pb.MoveResponse{Success: false, Msg: "Internal Server Error"},
			status.Errorf(codes.Internal, "failed to publish: %v", err)
	}
	return &pb.MoveResponse{Success: true, Msg: "Success"}, nil
}

func NewIngestServer(mp broker.MovePublisher, mc cache.MatchStateCache) *IngestServer {
	return &IngestServer{publisher: mp, mCache: mc}
}
