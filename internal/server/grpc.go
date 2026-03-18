package server

import (
	"context"
	"log"

	"github.com/1saswata/chess-broadcast-engine/internal/broker"
	"github.com/1saswata/chess-broadcast-engine/internal/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type IngestServer struct {
	pb.UnimplementedChessIngestServiceServer
	publisher broker.MovePublisher
}

func (is *IngestServer) RecordMove(ctx context.Context, m *pb.Move) (*pb.MoveResponse, error) {
	log.Print(m.String())
	b, err := proto.Marshal(m)
	if err != nil {
		log.Println("Unable to serialize - ", err.Error())
		return &pb.MoveResponse{Success: false, Msg: "Internal Server Error"},
			status.Errorf(codes.Internal, "failed to publish: %v", err)
	}
	if err = is.publisher.PublishMove(ctx, b); err != nil {
		log.Println("Unable to serialize - ", err.Error())
		return &pb.MoveResponse{Success: false, Msg: "Internal Server Error"},
			status.Errorf(codes.Internal, "failed to publish: %v", err)
	}
	return &pb.MoveResponse{Success: true, Msg: "Success"}, nil
}

func NewIngestServer(mp broker.MovePublisher) *IngestServer {
	return &IngestServer{publisher: mp}
}
