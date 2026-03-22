package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in dev; restrict in production
	},
}

// Client represents a single WebSocket connection
type Client struct {
	Conn    *websocket.Conn
	OrderID string
	Send    chan []byte
}

// WSHub maintains the set of active clients and broadcasts messages
type WSHub struct {
	mu         sync.RWMutex
	clients    map[string]map[*Client]bool // orderID -> set of clients
	register   chan *Client
	unregister chan *Client
	broadcast  chan WSMessage
}

// WSMessage contains a location update to broadcast
type WSMessage struct {
	OrderID string  `json:"order_id"`
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	Status  string  `json:"status"`
}

var Hub *WSHub

// InitWebSocketHub creates and starts the WebSocket hub
func InitWebSocketHub() {
	Hub = &WSHub{
		clients:    make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan WSMessage),
	}
	go Hub.run()
	fmt.Println("✅ WebSocket hub initialized")
}

func (h *WSHub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.OrderID] == nil {
				h.clients[client.OrderID] = make(map[*Client]bool)
			}
			h.clients[client.OrderID][client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.OrderID]; ok {
				if _, exists := clients[client]; exists {
					delete(clients, client)
					close(client.Send)
				}
				if len(clients) == 0 {
					delete(h.clients, client.OrderID)
				}
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.RLock()
			if clients, ok := h.clients[msg.OrderID]; ok {
				data, _ := json.Marshal(msg)
				for client := range clients {
					select {
					case client.Send <- data:
					default:
						close(client.Send)
						delete(clients, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastLocation sends a location update to all clients tracking a specific order
func BroadcastLocation(orderID string, lat, lng float64, status string) {
	if Hub == nil {
		return
	}
	Hub.broadcast <- WSMessage{
		OrderID: orderID,
		Lat:     lat,
		Lng:     lng,
		Status:  status,
	}
}

// HandleWebSocket upgrades HTTP to WebSocket for real-time delivery tracking
// GET /ws/track/:orderId
func HandleWebSocket(c *gin.Context) {
	orderID := c.Param("orderId")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("WebSocket upgrade error:", err)
		return
	}

	client := &Client{
		Conn:    conn,
		OrderID: orderID,
		Send:    make(chan []byte, 256),
	}

	Hub.register <- client

	// Writer goroutine: sends messages to the client
	go func() {
		defer func() {
			Hub.unregister <- client
			conn.Close()
		}()

		for message := range client.Send {
			err := conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				break
			}
		}
	}()

	// Reader goroutine: reads pings/messages from client (keeps connection alive)
	go func() {
		defer func() {
			Hub.unregister <- client
			conn.Close()
		}()

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}()
}
