package hub

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// Constants for WebSocket timeouts and message limits.
const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

// Upgrader is used to upgrade an HTTP connection to a WebSocket connection.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all connections for simplicity. Adjust as needed.
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Hub maintains the set of active clients and broadcasts messages to them.
type Hub struct {
	// Registered clients.
	Clients map[*Client]bool

	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client

	// Event Callbacks
	OnConnect    func(client *Client)
	OnDisconnect func(client *Client)
	OnMessage    func(client *Client, message []byte)
}

// NewHub creates a new Hub instance.
func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

// Run starts the hub's main loop. It listens for register, unregister, and broadcast events.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client] = true

			log.Printf("websocket: client (%s) connected; total clients: %d", client.RemoteAddr, len(h.Clients))

			if h.OnConnect != nil {
				h.OnConnect(client)
			}
		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)

				log.Printf("websocket: client (%s) disconnected; total clients: %d", client.RemoteAddr, len(h.Clients))

				if h.OnDisconnect != nil {
					h.OnDisconnect(client)
				}
			}
		case message := <-h.Broadcast:
			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}
		}
	}
}

// unregister removes a client; helper for client disconnect.
func (h *Hub) unregister(client *Client) {
	h.Unregister <- client
}

// ServeWs handles WebSocket requests from the peer.
// It upgrades the HTTP connection to a WebSocket, creates a client, and registers it with the hub.
func (h *Hub) ServeWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("websocket: upgrade error:", err)
		return
	}

	client := &Client{
		RemoteAddr: r.RemoteAddr,
		Hub:        h,
		Conn:       conn,
		Send:       make(chan []byte, maxMessageSize),
	}
	h.Register <- client

	go client.writePump()
	go client.readPump()
}
