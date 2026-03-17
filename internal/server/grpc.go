package server

import (
	"context"
	"log"

	"github.com/1saswata/chess-broadcast-engine/internal/pb"
)

type IngestServer struct {
	pb.UnimplementedChessIngestServiceServer
}

func (is *IngestServer) RecordMove(ctx context.Context, m *pb.Move) (*pb.MoveResponse, error) {
	log.Print(m.String())
	return &pb.MoveResponse{Success: true, Msg: "Success"}, nil
}

func NewIngestServer() *IngestServer {
	return &IngestServer{}
}
