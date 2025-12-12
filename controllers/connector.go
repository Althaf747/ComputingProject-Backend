package controllers

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ProxyWebcam(c *gin.Context) {
	resp, err := http.Get("http://localhost:5000/api/stream")
	if err != nil {
		c.Status(http.StatusBadGateway)
		return
	}
	c.Header("Content-Type", resp.Header.Get("Content-Type"))
	io.Copy(c.Writer, resp.Body)
}
