package controllers

import (
	"bufio"
	"bytes"
	"comproBackend/config"
	"comproBackend/models"
	"comproBackend/utils"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const pythonBaseURL = "http://localhost:5000"

func ProxyCameraStream(c *gin.Context) {
	resp, err := http.Get(pythonBaseURL + "/api/camera/stream")
	if err != nil {
		log.Println("Camera stream error:", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to connect to camera service"})
		return
	}
	defer resp.Body.Close()

	c.Header("Content-Type", resp.Header.Get("Content-Type"))
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	io.Copy(c.Writer, resp.Body)
}

func ProxyCameraSnapshot(c *gin.Context) {
	resp, err := http.Get(pythonBaseURL + "/api/camera/snapshot")
	if err != nil {
		log.Println("Camera snapshot error:", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to get snapshot"})
		return
	}
	defer resp.Body.Close()

	c.Header("Content-Type", resp.Header.Get("Content-Type"))
	io.Copy(c.Writer, resp.Body)
}

func ProxyCameraEvents(c *gin.Context) {
	resp, err := http.Get(pythonBaseURL + "/api/camera/events")
	if err != nil {
		log.Println("Camera events error:", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to connect to camera events"})
		return
	}
	defer resp.Body.Close()

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Streaming not supported"})
		return
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				log.Println("SSE read error:", err)
			}
			break
		}

		// Write to SSE client
		c.Writer.Write(line)
		flusher.Flush()

		// Parse and broadcast detection events to WebSocket clients
		lineStr := string(line)
		if strings.HasPrefix(lineStr, "data: ") {
			jsonData := strings.TrimPrefix(lineStr, "data: ")
			jsonData = strings.TrimSpace(jsonData)

			var event map[string]interface{}
			if err := json.Unmarshal([]byte(jsonData), &event); err == nil {
				// Broadcast to WebSocket clients
				utils.BroadcastEvent(event)

				// If it's a detection event, save to database
				if eventType, ok := event["type"].(string); ok && eventType == "detection" {
					go saveDetectionLog(event)
				}
			}
		}
	}
}

func saveDetectionLog(event map[string]interface{}) {
	data, ok := event["data"].(map[string]interface{})
	if !ok {
		return
	}

	name, _ := data["name"].(string)
	if name == "" || name == "Unknown" {
		return // Don't save unknown detections
	}

	authorized, _ := data["authorized"].(bool)
	confidence, _ := data["confidence"].(float64)
	role, _ := data["role"].(string)
	timestamp, _ := data["timestamp"].(string)

	if timestamp == "" {
		timestamp = time.Now().Format(time.RFC3339)
	}
	if role == "" {
		role = "unknown"
	}

	logEntry := models.Log{
		Authorized: authorized,
		Confidence: confidence,
		Name:       name,
		Role:       role,
		Timestamp:  timestamp,
	}

	if err := config.DB.Create(&logEntry).Error; err != nil {
		log.Println("Failed to save detection log:", err)
	} else {
		log.Printf("Detection logged: %s (authorized: %v, confidence: %.2f)", name, authorized, confidence)
	}
}

func GetCameraStatus(c *gin.Context) {
	resp, err := http.Get(pythonBaseURL + "/api/camera/status")
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to get camera status"})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	c.JSON(resp.StatusCode, result)
}

func StartCamera(c *gin.Context) {
	resp, err := http.Post(pythonBaseURL+"/api/camera/start", "application/json", nil)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to start camera"})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	c.JSON(resp.StatusCode, result)
}

func StopCamera(c *gin.Context) {
	resp, err := http.Post(pythonBaseURL+"/api/camera/stop", "application/json", nil)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to stop camera"})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	c.JSON(resp.StatusCode, result)
}

func GetCameraConfig(c *gin.Context) {
	resp, err := http.Get(pythonBaseURL + "/api/camera/config")
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to get camera config"})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	c.JSON(resp.StatusCode, result)
}

func UpdateCameraConfig(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	resp, err := http.Post(pythonBaseURL+"/api/camera/config", "application/json", bytes.NewReader(body))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to update camera config"})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	c.JSON(resp.StatusCode, result)
}

func GetCameraZones(c *gin.Context) {
	resp, err := http.Get(pythonBaseURL + "/api/camera/zones")
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to get camera zones"})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	c.JSON(resp.StatusCode, result)
}

func UpdateCameraZones(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	resp, err := http.Post(pythonBaseURL+"/api/camera/zones", "application/json", bytes.NewReader(body))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to update camera zones"})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	c.JSON(resp.StatusCode, result)
}

func SetupDroidCam(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	resp, err := http.Post(pythonBaseURL+"/api/camera/test/droidcam", "application/json", bytes.NewReader(body))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to setup DroidCam"})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	c.JSON(resp.StatusCode, result)
}

func SetupRTSP(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	resp, err := http.Post(pythonBaseURL+"/api/camera/test/rtsp", "application/json", bytes.NewReader(body))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to setup RTSP"})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	c.JSON(resp.StatusCode, result)
}
