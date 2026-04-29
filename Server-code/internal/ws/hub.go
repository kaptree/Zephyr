package ws

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"labelpro-server/internal/utils"

	"labelpro-server/internal/logger"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Hub struct {
	rooms      map[string]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan *Message
	mu         sync.RWMutex
}

type Client struct {
	ID     string
	Name   string
	RoomID string
	Conn   *websocket.Conn
	Send   chan []byte
}

type Message struct {
	RoomID  string
	Data    []byte
	Exclude *Client
}

var DefaultHub *Hub

func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Message),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if _, ok := h.rooms[client.RoomID]; !ok {
				h.rooms[client.RoomID] = make(map[*Client]bool)
			}
			h.rooms[client.RoomID][client] = true
			h.mu.Unlock()
			h.broadcastPresence(client.RoomID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.rooms[client.RoomID]; ok {
				delete(h.rooms[client.RoomID], client)
				if len(h.rooms[client.RoomID]) == 0 {
					delete(h.rooms, client.RoomID)
				}
			}
			h.mu.Unlock()
			close(client.Send)
			h.broadcastPresence(client.RoomID)

		case msg := <-h.broadcast:
			h.mu.RLock()
			if clients, ok := h.rooms[msg.RoomID]; ok {
				for client := range clients {
					if client != msg.Exclude {
						select {
						case client.Send <- msg.Data:
						default:
							close(client.Send)
							delete(h.rooms[msg.RoomID], client)
						}
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) broadcastPresence(roomID string) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.rooms[roomID]
	if !ok {
		return
	}

	var onlineUsers []map[string]string
	for client := range clients {
		onlineUsers = append(onlineUsers, map[string]string{
			"user_id": client.ID,
			"name":    client.Name,
		})
	}

	presenceData, _ := json.Marshal(map[string]interface{}{
		"event":        "presence:update",
		"online_users": onlineUsers,
	})

	for client := range clients {
		select {
		case client.Send <- presenceData:
		default:
		}
	}
}

func HandleWebSocket(hub *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		noteID := c.Param("note_id")
		token := c.Query("token")

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}

		claims, err := utils.ParseToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			logger.Error("websocket upgrade failed", zap.Error(err))
			return
		}

		client := &Client{
			ID:     claims.UserID,
			Name:   claims.Username,
			RoomID: noteID,
			Conn:   conn,
			Send:   make(chan []byte, 256),
		}

		hub.register <- client

		go client.writePump()
		go client.readPump(hub)
	}
}

func (c *Client) readPump(hub *Hub) {
	defer func() {
		hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512 * 1024)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		var event map[string]interface{}
		if err := json.Unmarshal(message, &event); err != nil {
			continue
		}

		eventType, _ := event["event"].(string)
		switch eventType {
		case "room:join":
			hub.broadcastPresence(c.RoomID)
		case "canvas:update":
			syncData, _ := json.Marshal(map[string]interface{}{
				"event":      "canvas:sync",
				"column_id":  event["column_id"],
				"content":    event["content"],
				"updated_by": c.Name,
				"version":    event["version"],
			})
			hub.broadcast <- &Message{
				RoomID:  c.RoomID,
				Data:    syncData,
				Exclude: c,
			}
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func InitHub() *Hub {
	hub := NewHub()
	DefaultHub = hub
	go hub.Run()
	return hub
}
