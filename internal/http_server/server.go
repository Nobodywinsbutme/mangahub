package http_server

import (
	"log"

	"github.com/Nobodywinsbutme/mangahub/internal/auth"
	"github.com/gin-gonic/gin"
)

func Start(port string) {
	r := gin.Default()

	// Register route groups
	auth.RegisterRoutes(r)

	// Health check endpoint — useful for testing
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	log.Printf("✓ HTTP API Server starting on http://localhost:%s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}
