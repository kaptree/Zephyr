package ws

import (
	"encoding/json"
	"net/http"
	"strings"
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
	ID            string
	Name          string
	RoomID        string
	Conn          *websocket.Conn
	Send          chan []byte
	EditingNoteID string
	// TokenExpiresAt 建立连接时所携带 token 的过期时间（零值表示不过期校验）。
	// 修复：cookie(token) 过期后仍显示在线的 bug —— Hub 定期扫描，过期连接强制断开。
	TokenExpiresAt time.Time
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
		// 需求30：带缓冲，避免连续推送（如 notification:new + note:updated）时非阻塞发送被静默丢弃
		broadcast: make(chan *Message, 256),
	}
}

func (h *Hub) Run() {
	// 修复：定期清理 token 已过期的连接，避免 cookie 过期后仍显示在线
	expireTicker := time.NewTicker(30 * time.Second)
	defer expireTicker.Stop()

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
			h.BroadcastPresence()

		case client := <-h.unregister:
			h.mu.Lock()
			if client.EditingNoteID != "" {
				idleData, _ := json.Marshal(map[string]interface{}{
					"event":   "note:idle",
					"note_id": client.EditingNoteID,
					"user_id": client.ID,
					"name":    client.Name,
				})
				if clients, ok := h.rooms[client.RoomID]; ok {
					for c := range clients {
						if c != client {
							select {
							case c.Send <- idleData:
							default:
							}
						}
					}
				}
			}
			if _, ok := h.rooms[client.RoomID]; ok {
				delete(h.rooms[client.RoomID], client)
				if len(h.rooms[client.RoomID]) == 0 {
					delete(h.rooms, client.RoomID)
				}
			}
			h.mu.Unlock()
			close(client.Send)
			h.broadcastPresence(client.RoomID)
			h.BroadcastPresence()

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
		case <-expireTicker.C:
			h.cleanupExpired()
		}
	}
}

// cleanupExpired 扫描所有房间，强制断开 token 已过期的连接并重新广播在线列表。
// 修复：用户 cookie(token) 过期后，只要 WebSocket 连接还挂着，其他用户仍能看到其在线。
func (h *Hub) cleanupExpired() {
	now := time.Now()

	h.mu.Lock()
	for roomID, clients := range h.rooms {
		for client := range clients {
			if !client.TokenExpiresAt.IsZero() && now.After(client.TokenExpiresAt) {
				// 先通过 Send 队列告知客户端 token 已过期（writePump 单写者，并发安全），
				// 稍等 writePump 把 auth:expired 发送出去后再关闭连接，确保客户端能收到明确事件。
				// readPump 读到错误后会自动走 unregister 流程移除该连接并广播新在线列表。
				expiredData, _ := json.Marshal(map[string]interface{}{"event": "auth:expired"})
				select {
				case client.Send <- expiredData:
				default:
				}
				go func(conn *websocket.Conn) {
					time.Sleep(500 * time.Millisecond)
					conn.Close()
				}(client.Conn)
				delete(clients, client)
			}
		}
		if len(clients) == 0 {
			delete(h.rooms, roomID)
		}
	}
	h.mu.Unlock()

	h.BroadcastPresence()
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

// PushToUser 向指定用户的个人通知通道推送数据（非阻塞）
func (h *Hub) PushToUser(userID string, data []byte) {
	if h == nil {
		return
	}
	select {
	case h.broadcast <- &Message{RoomID: "user:" + userID, Data: data}:
	default:
	}
}

// BroadcastToRoom 需求29：向指定房间（note）内的所有在线客户端广播（协同指令实时下发）
func (h *Hub) BroadcastToRoom(roomID string, data []byte) {
	if h == nil {
		return
	}
	select {
	case h.broadcast <- &Message{RoomID: roomID, Data: data}:
	default:
	}
}

// IsUserOnline 判断用户是否有活跃的个人 WebSocket 连接
func (h *Hub) IsUserOnline(userID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	clients, ok := h.rooms["user:"+userID]
	return ok && len(clients) > 0
}

// OnlineUserIDs 返回当前所有在线用户的 ID 列表
func (h *Hub) OnlineUserIDs() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	var ids []string
	for roomID, clients := range h.rooms {
		if strings.HasPrefix(roomID, "user:") && len(clients) > 0 {
			ids = append(ids, strings.TrimPrefix(roomID, "user:"))
		}
	}
	return ids
}

// BroadcastPresence 向所有在线用户广播在线用户列表（用户上线/下线时调用）
func (h *Hub) BroadcastPresence() {
	if h == nil {
		return
	}
	ids := h.OnlineUserIDs()
	if ids == nil {
		ids = []string{}
	}
	presenceData, _ := json.Marshal(map[string]interface{}{
		"event":      "presence:update",
		"online_ids": ids,
	})

	h.mu.RLock()
	defer h.mu.RUnlock()
	for roomID, clients := range h.rooms {
		if !strings.HasPrefix(roomID, "user:") {
			continue
		}
		for client := range clients {
			select {
			case client.Send <- presenceData:
			default:
			}
		}
	}
}

// HandleUserWebSocket 个人通知通道：GET /ws/user/:user_id?token=xxx
func HandleUserWebSocket(hub *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
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
			ID:             claims.UserID,
			Name:           claims.Username,
			RoomID:         "user:" + claims.UserID,
			Conn:           conn,
			Send:           make(chan []byte, 256),
			TokenExpiresAt: tokenExpiry(claims),
		}

		hub.register <- client

		go client.writePump()
		go client.readPump(hub)
	}
}

// tokenExpiry 从 JWT claims 中提取 token 过期时间（缺失时为零值，表示不做过期校验）
func tokenExpiry(claims *utils.Claims) time.Time {
	if claims != nil && claims.ExpiresAt != nil {
		return claims.ExpiresAt.Time
	}
	return time.Time{}
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
			ID:             claims.UserID,
			Name:           claims.Username,
			RoomID:         noteID,
			Conn:           conn,
			Send:           make(chan []byte, 256),
			TokenExpiresAt: tokenExpiry(claims),
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
		case "note:editing":
			c.EditingNoteID, _ = event["note_id"].(string)
			editData, _ := json.Marshal(map[string]interface{}{
				"event":   "note:editing",
				"note_id": c.EditingNoteID,
				"user_id": c.ID,
				"name":    c.Name,
			})
			hub.broadcast <- &Message{
				RoomID:  c.RoomID,
				Data:    editData,
				Exclude: c,
			}
		case "note:idle":
			editID := c.EditingNoteID
			c.EditingNoteID = ""
			idleData, _ := json.Marshal(map[string]interface{}{
				"event":   "note:idle",
				"note_id": event["note_id"],
				"user_id": c.ID,
				"name":    c.Name,
			})
			hub.broadcast <- &Message{
				RoomID:  c.RoomID,
				Data:    idleData,
				Exclude: c,
			}
			_ = editID
		case "note:updated":
			updData, _ := json.Marshal(map[string]interface{}{
				"event":   "note:updated",
				"note_id": event["note_id"],
				"action":  event["action"],
				"user_id": c.ID,
				"name":    c.Name,
			})
			hub.broadcast <- &Message{
				RoomID:  c.RoomID,
				Data:    updData,
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

func HandleGroupWebSocket(hub *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupID := c.Param("group_id")
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
			ID:             claims.UserID,
			Name:           claims.Username,
			RoomID:         "group:" + groupID,
			Conn:           conn,
			Send:           make(chan []byte, 256),
			TokenExpiresAt: tokenExpiry(claims),
		}

		hub.register <- client

		go client.writePump()
		go client.readPump(hub)
	}
}
