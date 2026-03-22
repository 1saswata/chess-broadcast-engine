package websocket

import (
	"sync"
)

type Hub struct {
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	clients    map[*Client]bool
	sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (hub *Hub) Run() {
	for {
		select {
		case client := <-hub.register:
			hub.Lock()
			hub.clients[client] = true
			hub.Unlock()
		case client := <-hub.unregister:
			hub.Lock()
			delete(hub.clients, client)
			hub.Unlock()
		case msg := <-hub.broadcast:
			hub.RLock()
			for client := range hub.clients {
				client.send <- msg
			}
			hub.RUnlock()
		}
	}
}

func (hub *Hub) Broadcast(m []byte) {
	hub.broadcast <- m
}
