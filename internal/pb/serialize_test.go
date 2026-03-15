package pb

import (
	"testing"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestMoveProtoMarshal(t *testing.T) {
	tests := []struct {
		name    string
		input   *Move
		payload []byte
		wantErr bool
	}{
		{
			name: "Happy Path",
			input: &Move{MatchId: 1, CurrentPlayer: Player_PLAYER_WHITE,
				StartingSquare: "C4", DestinationSquare: "C5",
				Timestamp: timestamppb.New(time.Now())},
			payload: nil,
			wantErr: false,
		},
		{
			name: "Unhappy Path - bad payload",
			input: &Move{MatchId: 1, CurrentPlayer: Player_PLAYER_WHITE,
				StartingSquare: "C4", DestinationSquare: "C5",
				Timestamp: timestamppb.New(time.Now())},
			payload: []byte("invalid binary data"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if tt.payload == nil {
				tt.payload, err = proto.Marshal(tt.input)
				if err != nil {
					t.Error("Err: ", err.Error())
				}
			}
			var newMove Move
			err = proto.Unmarshal(tt.payload, &newMove)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Want - Error, Got - %s", newMove.String())
				}
			} else {
				if err != nil {
					t.Error("Err: ", err.Error())
				}
				if !proto.Equal(tt.input, &newMove) {
					t.Errorf("Want: %v , Got: %v", tt.input.String(), newMove.String())
				}
			}
		})
	}
}
