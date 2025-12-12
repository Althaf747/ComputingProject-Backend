package utils

import (
	"comproBackend/config"
	"comproBackend/models"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func WsHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	log.Println("Client connected")

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
