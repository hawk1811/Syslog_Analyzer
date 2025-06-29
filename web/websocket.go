package web

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketManager manages WebSocket connections
type WebSocketManager struct {
	clients    map[*Client]bool
	clientsMux sync.RWMutex
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	upgrader   websocket.Upgrader
	running    bool
	stopChan   chan bool
}

// Client represents a WebSocket client connection
type Client struct {
	conn     *websocket.Conn
	send     chan []byte
	manager  *WebSocketManager
	lastPing time.Time
	id       string
}

// NewWebSocketManager creates a new WebSocket manager
func NewWebSocketManager() *WebSocketManager {
	return &WebSocketManager{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		stopChan:   make(chan bool),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// Allow connections from any origin in development
				// In production, you should validate the origin
				return true
			},
		},
	}
}

// Start starts the WebSocket manager
func (wsm *WebSocketManager) Start() {
	wsm.running = true
	log.Printf("🔌 WebSocket manager started")
	
	// Start ping ticker to keep connections alive
	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()
	
	for wsm.running {
		select {
		case client := <-wsm.register:
			wsm.registerClient(client)
			
		case client := <-wsm.unregister:
			wsm.unregisterClient(client)
			
		case message := <-wsm.broadcast:
			wsm.broadcastMessage(message)
			
		case <-pingTicker.C:
			wsm.pingClients()
			
		case <-wsm.stopChan:
			log.Printf("🔌 WebSocket manager stopping...")
			return
		}
	}
}

// Stop stops the WebSocket manager
func (wsm *WebSocketManager) Stop() {
	if !wsm.running {
		return
	}
	
	wsm.running = false
	
	// Send stop signal
	select {
	case wsm.stopChan <- true:
	default:
	}
	
	// Close all client connections
	wsm.clientsMux.Lock()
	for client := range wsm.clients {
		close(client.send)
		client.conn.Close()
		delete(wsm.clients, client)
	}
	wsm.clientsMux.Unlock()
	
	log.Printf("✓ WebSocket manager stopped")
}

// Broadcast sends data to all connected clients
func (wsm *WebSocketManager) Broadcast(data interface{}) {
	if !wsm.running {
		return
	}
	
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("⚠ Error marshaling WebSocket data: %v", err)
		return
	}
	
	select {
	case wsm.broadcast <- jsonData:
	default:
		// Channel is full, skip this broadcast
		log.Printf("⚠ WebSocket broadcast channel full, skipping message")
	}
}

// HandleWebSocket handles WebSocket upgrade requests - FIXED METHOD NAME
func (wsm *WebSocketManager) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	log.Printf("🔌 WebSocket upgrade request from %s", r.RemoteAddr)
	
	// Perform WebSocket upgrade
	conn, err := wsm.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ WebSocket upgrade error: %v", err)
		return
	}
	
	log.Printf("✅ WebSocket connection established with %s", r.RemoteAddr)
	
	// Create new client
	client := &Client{
		conn:     conn,
		send:     make(chan []byte, 256),
		manager:  wsm,
		lastPing: time.Now(),
		id:       r.RemoteAddr,
	}
	
	// Register the client
	wsm.register <- client
	
	// Start goroutines for this client
	go client.writePump()
	go client.readPump()
}

// registerClient registers a new client
func (wsm *WebSocketManager) registerClient(client *Client) {
	wsm.clientsMux.Lock()
	wsm.clients[client] = true
	clientCount := len(wsm.clients)
	wsm.clientsMux.Unlock()
	
	log.Printf("✅ WebSocket client connected from %s (total: %d)", client.id, clientCount)
}

// unregisterClient unregisters a client
func (wsm *WebSocketManager) unregisterClient(client *Client) {
	wsm.clientsMux.Lock()
	if _, ok := wsm.clients[client]; ok {
		delete(wsm.clients, client)
		close(client.send)
		clientCount := len(wsm.clients)
		wsm.clientsMux.Unlock()
		log.Printf("❌ WebSocket client disconnected from %s (total: %d)", client.id, clientCount)
	} else {
		wsm.clientsMux.Unlock()
	}
}

// broadcastMessage sends a message to all clients
func (wsm *WebSocketManager) broadcastMessage(message []byte) {
	wsm.clientsMux.RLock()
	defer wsm.clientsMux.RUnlock()
	
	for client := range wsm.clients {
		select {
		case client.send <- message:
		default:
			// Client's send channel is full, remove the client
			delete(wsm.clients, client)
			close(client.send)
			log.Printf("⚠ Removed unresponsive WebSocket client %s", client.id)
		}
	}
}

// pingClients sends ping messages to all clients
func (wsm *WebSocketManager) pingClients() {
	wsm.clientsMux.RLock()
	clients := make([]*Client, 0, len(wsm.clients))
	for client := range wsm.clients {
		clients = append(clients, client)
	}
	wsm.clientsMux.RUnlock()
	
	for _, client := range clients {
		// Check if client hasn't responded to ping in a while
		if time.Since(client.lastPing) > 60*time.Second {
			// Client is considered dead, remove it
			wsm.unregister <- client
			continue
		}
		
		// Send ping
		client.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
			wsm.unregister <- client
		}
	}
}

// readPump pumps messages from the WebSocket connection to the manager
func (c *Client) readPump() {
	defer func() {
		c.manager.unregister <- c
		c.conn.Close()
	}()
	
	// Set read deadline and pong handler
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.lastPing = time.Now()
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("⚠ WebSocket error from %s: %v", c.id, err)
			}
			break
		}
		
		// Handle incoming messages if needed
		log.Printf("📨 WebSocket message from %s: %s", c.id, string(message))
	}
}

// writePump pumps messages from the manager to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// The manager closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			
			// Add queued messages to the current WebSocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}
			
			if err := w.Close(); err != nil {
				return
			}
			
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// GetClientCount returns the number of connected clients
func (wsm *WebSocketManager) GetClientCount() int {
	wsm.clientsMux.RLock()
	defer wsm.clientsMux.RUnlock()
	return len(wsm.clients)
}