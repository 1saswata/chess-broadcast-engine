package server

import (
	"context"
	"log"
	"net"
	"testing"

	"github.com/1saswata/chess-broadcast-engine/internal/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const bufSize = 1024 * 1024

var lis = bufconn.Listen(bufSize)

func bufDialer(ctx context.Context, target string) (net.Conn, error) {
	return lis.Dial()
}
func TestMove(t *testing.T) {
	s := grpc.NewServer()
	pb.RegisterChessIngestServiceServer(s, IngestServer{})
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()

	conn, err := grpc.NewClient("passthrough:///bufnet", grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := pb.NewChessIngestServiceClient(conn)
	t.Run("Happy Path", func(t *testing.T) {
		mr, err := client.RecordMove(context.Background(), &pb.Move{
			MatchId:           1,
			CurrentPlayer:     pb.Player_PLAYER_WHITE,
			StartingSquare:    "C5",
			DestinationSquare: "C6",
			Timestamp:         timestamppb.Now(),
		})
		if mr.Success != true || err != nil {
			t.Error("Want - Success Got - ", mr.String())
		}
	})
}
