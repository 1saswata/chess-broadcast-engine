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

type MockPublisher struct {
}

func (m MockPublisher) PublishMove(ctx context.Context, move []byte) error {
	return nil
}

type MockCache struct {
}

func (m MockCache) AppendMove(ctx context.Context, matchID int32, move []byte) error {
	return nil
}

func (m MockCache) GetMoveHistory(ctx context.Context, matchID int32) ([][]byte, error) {
	return nil, nil
}

func TestMove(t *testing.T) {
	var lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	pb.RegisterChessIngestServiceServer(s, NewIngestServer(MockPublisher{}, MockCache{}))
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()
	defer s.GracefulStop()
	conn, err := grpc.NewClient("passthrough:///bufnet", grpc.WithContextDialer(
		func(ctx context.Context, target string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := pb.NewChessIngestServiceClient(conn)

	tests := []struct {
		name string
		move *pb.Move
		want bool
	}{
		{
			name: "Happy Path",
			move: &pb.Move{
				MatchId:           1,
				CurrentPlayer:     pb.Player_PLAYER_WHITE,
				StartingSquare:    "C5",
				DestinationSquare: "C6",
				Timestamp:         timestamppb.Now(),
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr, err := client.RecordMove(context.Background(), tt.move)
			if err != nil {
				t.Error("Error - ", err.Error())
			}
			if mr.Success != tt.want {
				t.Errorf("Want %t Got %t", tt.want, mr.Success)
			}
		})
	}
}
