package controllers

import (
	"comproBackend/config"
	"comproBackend/models"
	"comproBackend/services"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func GetLogs(c *gin.Context) {
	var Logs []models.Log

	if err := config.DB.Find(&Logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": Logs})
}

func GetFilteredLogs(c *gin.Context) {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load timezone"})
		return
	}

	name := c.Query("name")
	period := c.Query("period")
	dateStr := c.Query("date")
	startDateStr := c.Query("start")
	endDateStr := c.Query("end")

	var startStr, endStr string
	var startLocal, endLocal time.Time
	now := time.Now().In(loc)

	switch period {
	case "range":
		if startDateStr == "" || endDateStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Start and end date parameters are required for 'range' period"})
			return
		}
		startParsed, err := time.ParseInLocation("2006-01-02", startDateStr, loc)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format. Use YYYY-MM-DD"})
			return
		}
		endParsed, err := time.ParseInLocation("2006-01-02", endDateStr, loc)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format. Use YYYY-MM-DD"})
			return
		}
		startLocal = time.Date(startParsed.Year(), startParsed.Month(), startParsed.Day(), 0, 0, 0, 0, loc)
		endLocal = time.Date(endParsed.Year(), endParsed.Month(), endParsed.Day(), 23, 59, 59, 0, loc).Add(time.Second)
	case "date":
		if dateStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Date parameter is required for 'date' period"})
			return
		}
		parsedDate, err := time.ParseInLocation("2006-01-02", dateStr, loc)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD"})
			return
		}
		startLocal = time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, loc)
		endLocal = time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 23, 59, 59, 0, loc).Add(time.Second)
	case "today":
		fallthrough
	default:
		startLocal = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
		endLocal = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, loc).Add(time.Second)
	}

	startStr = startLocal.Format("2006-01-02T15:04:05")
	endStr = endLocal.Format("2006-01-02T15:04:05")

	var logs []models.Log

	query := config.DB.Where("timestamp >= ? AND timestamp <= ?", startStr, endStr)

	if name != "" {
		query = query.Where("name = ?", name)
	}

	if err := query.Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var UnkVisitors int
	for _, log := range logs {
		if log.Name == "Unknown" {
			UnkVisitors++
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": logs, "count": len(logs), "unknown_visitors": UnkVisitors})
}

func CreateLog(c *gin.Context) {
	var Log models.Log

	if err := c.ShouldBindJSON(&Log); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := config.DB.Create(&Log).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	go sendLogNotification(Log)

	c.JSON(http.StatusCreated, gin.H{"data": Log})
}

func sendLogNotification(log models.Log) {
	var users []models.User
	if err := config.DB.Where("fcm_token != '' AND fcm_token IS NOT NULL").Find(&users).Error; err != nil {
		fmt.Printf("Error fetching users for notification: %v\n", err)
		return
	}

	if len(users) == 0 {
		return
	}

	var title, body string
	if log.Authorized {
		title = "Access Granted"
		body = fmt.Sprintf("%s was detected at the door", log.Name)
	} else {
		title = "⚠️ Unknown Person Detected"
		body = "An unknown person was detected at the door"
	}

	data := map[string]string{
		"type":       "log",
		"log_id":     fmt.Sprintf("%d", log.ID),
		"name":       log.Name,
		"authorized": fmt.Sprintf("%t", log.Authorized),
		"timestamp":  log.Timestamp,
	}

	var tokens []string
	for _, user := range users {
		if user.FCMToken != "" {
			tokens = append(tokens, user.FCMToken)
		}
	}

	if len(tokens) > 0 {
		if err := services.SendPushNotificationToMultiple(tokens, title, body, data); err != nil {
			fmt.Printf("Error sending push notifications: %v\n", err)
		}
	}
}

func DeleteLog(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := config.DB.Where("id = ?", id).First(&models.Log{}).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Log not found"})
		return
	}

	if err := config.DB.Unscoped().Delete(&models.Log{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Log deleted successfully"})
}
