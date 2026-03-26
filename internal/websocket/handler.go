package websocket

import (
	"context"
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
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	matchID, err := strconv.Atoi(i)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	res, err := wh.rc.GetLatestMove(context.Background(), int32(matchID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	client := Client{
		hub:  wh.hub,
		conn: conn,
		send: make(chan []byte, 256),
	}
	if res != nil {
		var m pb.Move
		err := proto.Unmarshal(res, &m)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		res, err := protojson.Marshal(&m)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		client.send <- res
	}
	client.hub.register <- &client
	go client.writePump()
	go client.readPump()
}
