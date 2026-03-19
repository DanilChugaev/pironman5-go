package main

import (
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/gin-gonic/gin"
)

const httpPort = ":34001"

func setLED(r, g, b int) {
	exec.Command("sudo", "python3", "led_helper.py", fmt.Sprintf("%d", r), fmt.Sprintf("%d", g), fmt.Sprintf("%d", b)).Run()
}

func main() {
	fmt.Println("🚀 Pironman5-Go v0.8 (Pi 5 + Python LED helper)")

	// Тест при старте
	setLED(255, 0, 0)
	time.Sleep(4 * time.Second)
	setLED(0, 0, 0)

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "OK", "leds": "работают через SPI", "pi": "5"})
	})

	r.POST("/rgb", func(c *gin.Context) {
		color := c.Query("c")
		switch color {
		case "red":
			setLED(255, 0, 0)
		case "green":
			setLED(0, 255, 0)
		case "blue":
			setLED(0, 0, 255)
		case "white":
			setLED(255, 255, 255)
		default:
			setLED(0, 0, 0)
		}
		c.JSON(200, gin.H{"ok": true, "color": color})
	})

	log.Printf("Сервер на %s", httpPort)
	r.Run(httpPort)
}
