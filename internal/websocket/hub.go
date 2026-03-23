package websocket

type Hub struct {
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	clients    map[*Client]bool
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
			hub.clients[client] = true
		case client := <-hub.unregister:
			delete(hub.clients, client)
		case msg := <-hub.broadcast:
			for client := range hub.clients {
				select {
				case client.send <- msg:
				default:
					close(client.send)
					delete(hub.clients, client)
				}
			}
		}
	}
}

func (hub *Hub) Broadcast(m []byte) {
	hub.broadcast <- m
}
