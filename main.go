package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"

	"github.com/gin-gonic/gin"
)

const (
	// dataPin  = 10 // GPIO10 = BCM 10, физ. пин 19 (MOSI)
	// ledCount = 4
	httpPort = ":34001"
)

func printStatus() {
	cmd := exec.Command("sudo", "-E", "venv/bin/python3", "scripts/status/print_status.py")

	// Запуск и получение вывода
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Ошибка запуска: %s", err)
	}

	fmt.Println(string(output))
}

func main() {
	fmt.Println("🚀 Pironman5-Go v0.10 — go + python scripts")

	router := gin.Default()

	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true, "driver": "python script"})
	})

	router.GET("/api/status", func(c *gin.Context) {
		printStatus()
		c.JSON(http.StatusOK, gin.H{"success": true, "status": "?"})
	})

	router.POST("/api/rgb", func(c *gin.Context) {
		col := c.Query("c")

		c.JSON(http.StatusOK, gin.H{"success": true, "color": col})
	})

	log.Printf("Сервер на %s", httpPort)
	router.Run(httpPort)
}
