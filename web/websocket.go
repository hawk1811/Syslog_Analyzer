package web

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"syslog-analyzer/models"
)

// WSManager manages WebSocket connections and broadcasting
type WSManager struct {
	clients      map[*websocket.Conn]bool
	clientsMutex sync.RWMutex
	hub          chan []byte
	upgrader     websocket.Upgrader
}

// NewWSManager creates a new WebSocket manager
func NewWSManager() *WSManager {
	return &WSManager{
		clients: make(map[*websocket.Conn]bool),
		hub:     make(chan []byte, 256),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for simplicity
			},
		},
	}
}

// Start starts the WebSocket hub
func (ws *WSManager) Start() {
	go ws.runHub()
}

// HandleConnection handles a new WebSocket connection
func (ws *WSManager) HandleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()
	
	ws.clientsMutex.Lock()
	ws.clients[conn] = true
	ws.clientsMutex.Unlock()
	
	defer func() {
		ws.clientsMutex.Lock()
		delete(ws.clients, conn)
		ws.clientsMutex.Unlock()
	}()
	
	// Keep connection alive
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// BroadcastMetrics broadcasts metrics to all connected clients
func (ws *WSManager) BroadcastMetrics(sources []models.SourceMetrics, global models.GlobalMetrics) {
	response := map[string]interface{}{
		"sources": sources,
		"global":  global,
	}
	
	data, err := json.Marshal(response)
	if err != nil {
		return
	}
	
	select {
	case ws.hub <- data:
	default:
	}
}

// BroadcastRaw broadcasts raw JSON data to all connected clients
func (ws *WSManager) BroadcastRaw(data []byte) {
	select {
	case ws.hub <- data:
	default:
	}
}

// runHub manages the WebSocket message broadcasting
func (ws *WSManager) runHub() {
	for {
		select {
		case message := <-ws.hub:
			ws.clientsMutex.RLock()
			for client := range ws.clients {
				if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
					client.Close()
					delete(ws.clients, client)
				}
			}
			ws.clientsMutex.RUnlock()
		}
	}
}