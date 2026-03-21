package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/DanilChugaev/pironman5-go/pkg/status"
	"github.com/gin-gonic/gin"
)

const (
	// dataPin  = 10 // GPIO10 = BCM 10, физ. пин 19 (MOSI)
	// ledCount = 4
	httpPort = ":34001"
)

type ResponseDTO struct {
	Success bool   `json:"success"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func main() {
	fmt.Println("🚀 Pironman5-Go v0.10 — go + python scripts")

	router := gin.Default()

	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true, "driver": "python script"})
	})

	router.GET("/api/status", func(c *gin.Context) {
		// status.PrintStatus()
		statusObj := status.GetStatus()

		c.JSON(http.StatusOK, ResponseDTO{
			Success: true,
			Code:    http.StatusOK,
			Message: http.StatusText(http.StatusOK),
			Data:    statusObj,
		})
	})

	router.POST("/api/rgb", func(c *gin.Context) {
		col := c.Query("c")

		c.JSON(http.StatusOK, gin.H{"success": true, "color": col})
	})

	log.Printf("Сервер на %s", httpPort)
	router.Run(httpPort)
}
