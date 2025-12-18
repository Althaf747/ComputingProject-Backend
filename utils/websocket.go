package utils

import (
	"comproBackend/config"
	"comproBackend/models"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	conn *websocket.Conn
	send chan []byte
}

type ClientManager struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mutex      sync.RWMutex
}

var Manager = ClientManager{
	clients:    make(map[*Client]bool),
	broadcast:  make(chan []byte, 256),
	register:   make(chan *Client),
	unregister: make(chan *Client),
}

func (m *ClientManager) Start() {
	for {
		select {
		case client := <-m.register:
			m.mutex.Lock()
			m.clients[client] = true
			m.mutex.Unlock()
			log.Printf("Client registered. Total clients: %d", len(m.clients))

		case client := <-m.unregister:
			m.mutex.Lock()
			if _, ok := m.clients[client]; ok {
				delete(m.clients, client)
				close(client.send)
			}
			m.mutex.Unlock()
			log.Printf("Client unregistered. Total clients: %d", len(m.clients))

		case message := <-m.broadcast:
			m.mutex.RLock()
			for client := range m.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(m.clients, client)
				}
			}
			m.mutex.RUnlock()
		}
	}
}

func BroadcastEvent(event map[string]interface{}) {
	data, err := json.Marshal(event)
	if err != nil {
		log.Println("Failed to marshal event:", err)
		return
	}

	select {
	case Manager.broadcast <- data:
	default:
		log.Println("Broadcast channel full, dropping message")
	}
}

func WsHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	client := &Client{
		conn: conn,
		send: make(chan []byte, 256),
	}

	Manager.register <- client

	go client.writePump()

	log.Println("Client connected")

	defer func() {
		Manager.unregister <- client
		conn.Close()
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		log.Println("Received from client:", string(message))

		var logEntry models.Log
		if err := json.Unmarshal(message, &logEntry); err != nil {
			log.Println("JSON parse error:", err)

			_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"invalid JSON"}`))
			continue
		}

		if err := config.DB.Create(&logEntry).Error; err != nil {
			log.Println("DB error:", err)
			_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"db failed"}`))
			continue
		}

		_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"status":"saved"}`))
	}
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()

	for message := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Println("Write error:", err)
			return
		}
	}
}
