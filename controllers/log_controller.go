package controllers

import (
	"comproBackend/config"
	"comproBackend/models"
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
	path := c.FullPath()
	dateStr := c.Param("date")

	var startStr, endStr string
	var startLocal, endLocal time.Time
	now := time.Now().In(loc)

	if path == "/api/logs/last-7-days" {
		endLocal = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, loc).Add(time.Second)
		startLocal = endLocal.Add(-7 * 24 * time.Hour)
	} else if path == "/api/logs/last-month" {
		endLocal = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, loc).Add(time.Second)
		startLocal = endLocal.AddDate(0, -1, 0)
	} else if path == "/api/logs/today" || dateStr == "" {
		startLocal = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
		endLocal = startLocal.Add(24 * time.Hour)
	} else {
		parsedDate, err := time.ParseInLocation("2006-01-02", dateStr, loc)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD"})
			return
		}
		startLocal = time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, loc)
		endLocal = startLocal.Add(24 * time.Hour)
	}

	startStr = startLocal.Format("2006-01-02T15:04:05")
	endStr = endLocal.Format("2006-01-02T15:04:05")

	var logs []models.Log

	query := config.DB.Where("timestamp >= ? AND timestamp < ?", startStr, endStr)

	if name != "" {
		query = query.Where("name = ?", name)
	}

	if path == "/api/logs/last-7-days" || path == "/api/logs/last-month" {
		query = query.Order("timestamp DESC")
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

	c.JSON(http.StatusCreated, gin.H{"data": Log})
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
