package pb

import (
	"testing"

	"google.golang.org/protobuf/proto"
)

func TestMoveProto(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		m := Move{UMID: 1, CurrentPlayer: 1, StartingSquare: "C4", DestinationSquare: "C5"}
		b, err := proto.Marshal(&m)
		if err != nil {
			t.Error("Err: ", err.Error())
		}
		var newm Move
		err = proto.Unmarshal(b, &newm)
		if err != nil {
			t.Error("Err: ", err.Error())
		}
		if m.String() != newm.String() {
			t.Errorf("Want: %v , Got: %v", m.String(), newm.String())
		}
	})
	t.Run("Unhappy Path", func(t *testing.T) {
		b := []byte("invalid binary data")
		var m Move
		err := proto.Unmarshal(b, &m)
		if err == nil {
			t.Errorf("Want - Error, Got - %s", m.String())
		}
	})
}
