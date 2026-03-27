package websocket

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/1saswata/chess-broadcast-engine/internal/cache"
	"github.com/1saswata/chess-broadcast-engine/internal/pb"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type WsHandler struct {
	hub *Hub
	rc  cache.MatchStateCache
}

func NewWsHandler(h *Hub, rc cache.MatchStateCache) *WsHandler {
	return &WsHandler{hub: h, rc: rc}
}

func (wh *WsHandler) ServeHttp(w http.ResponseWriter, r *http.Request) {
	i := r.URL.Query().Get("match_id")
	matchID, err := strconv.Atoi(i)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	res, err := wh.rc.GetLatestMove(context.Background(), int32(matchID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	client := Client{
		matchID: int32(matchID),
		hub:     wh.hub,
		conn:    conn,
		send:    make(chan []byte, 256),
	}
	if res != nil {
		var m pb.Move
		err := proto.Unmarshal(res, &m)
		if err != nil {
			slog.Error("Error in cache move unmarshal", "Error", err.Error())
		} else {
			res, err := protojson.Marshal(&m)
			if err != nil {
				slog.Error("Error in cache json marshal", "Error", err.Error())
			} else {
				client.send <- res
			}
		}
	}
	client.hub.register <- &client
	go client.writePump()
	go client.readPump()
}
