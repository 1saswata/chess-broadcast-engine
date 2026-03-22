package websocket

import (
	"log/slog"
	"net/http"

	ws "github.com/gorilla/websocket"
)

var upgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type Client struct {
	hub  *Hub
	conn *ws.Conn
	send chan []byte
}

func (c *Client) readPump() {
	defer func() {
		err := c.conn.Close()
		if err != nil {
			slog.Error("Error closing connection", "Error", err)
		}
		c.hub.unregister <- c
	}()
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (c *Client) writePump() {
	defer func() {
		err := c.conn.Close()
		if err != nil {
			slog.Error("Error closing connection", "Error", err)
		}
		c.hub.unregister <- c
	}()
	for {
		msg := <-c.send
		err := c.conn.WriteMessage(ws.TextMessage, msg)
		if err != nil {
			slog.Error("Error sending message", "Error", err)
			break
		}
	}
}

func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	client := Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte),
	}
	client.hub.register <- &client
	go client.writePump()
	go client.readPump()
}
