package websocket

type Hub struct {
	broadcast  chan *BroadcastMessage
	register   chan *Client
	unregister chan *Client
	rooms      map[int32]map[*Client]bool
}

type BroadcastMessage struct {
	MatchID int32
	Payload []byte
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan *BroadcastMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		rooms:      make(map[int32]map[*Client]bool),
	}
}

func (hub *Hub) Run() {
	for {
		select {
		case client := <-hub.register:
			_, ok := hub.rooms[client.matchID]
			if !ok {
				hub.rooms[client.matchID] = make(map[*Client]bool)
			}
			hub.rooms[client.matchID][client] = true
		case client := <-hub.unregister:
			delete(hub.rooms[client.matchID], client)
			if len(hub.rooms[client.matchID]) == 0 {
				delete(hub.rooms, client.matchID)
			}
		case msg := <-hub.broadcast:
			for client := range hub.rooms[msg.MatchID] {
				select {
				case client.send <- msg.Payload:
				default:
					close(client.send)
					delete(hub.rooms[client.matchID], client)
					if len(hub.rooms[client.matchID]) == 0 {
						delete(hub.rooms, client.matchID)
					}
				}
			}
		}
	}
}

func (hub *Hub) Broadcast(m *BroadcastMessage) {
	hub.broadcast <- m
}
