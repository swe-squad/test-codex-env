package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// NewRouter builds the HTTP handler; ping2URL is the base URL of ping-2 (PING_2_URL).
func NewRouter(ping2URL string) *gin.Engine {
	client := &http.Client{Timeout: 3 * time.Second}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")
	api.GET("/ping", func(c *gin.Context) {
		req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet,
			strings.TrimRight(ping2URL, "/")+"/api/v1/ping", nil)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "ping-2 request failed"})
			return
		}
		res, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "ping-2 is unavailable"})
			return
		}
		defer res.Body.Close()

		var pong struct {
			Message string `json:"message"`
			Service string `json:"service"`
		}
		if res.StatusCode != http.StatusOK || json.NewDecoder(res.Body).Decode(&pong) != nil ||
			pong.Message != "pong" || pong.Service != "ping-2" {
			c.JSON(http.StatusBadGateway, gin.H{"error": "invalid response from ping-2"})
			return
		}
		c.JSON(http.StatusOK, pong)
	})

	return r
}
