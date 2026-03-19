package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	ws281x "github.com/rpi-ws281x/rpi-ws281x-go"
)

func main() {
	fmt.Println("🚀 Pironman5-Go тестовый сервис запущен")

	// ==================== WS281x ====================
	cfg := ws281x.DefaultConfig(4) // 4 светодиода
	cfg.Brightness = 64
	cfg.Channel = 0
	cfg.GpioPin = 10 // GPIO10 — правильный для Pironman5 + Pi 5 (SPI)

	strip, err := ws281x.MakeWS2811(&cfg)
	if err != nil {
		log.Fatalf("Не удалось создать strip: %v", err)
	}
	defer strip.Close()

	if err := strip.Init(); err != nil {
		log.Fatalf("Init не прошёл: %v", err)
	}

	// Тест: зажигаем все красным на 5 секунд при старте
	for i := 0; i < 4; i++ {
		strip.Leds(0)[i] = 0xFF0000 // красный
	}
	strip.Render()
	time.Sleep(5 * time.Second)
	for i := 0; i < 4; i++ {
		strip.Leds(0)[i] = 0x000000
	}
	strip.Render()

	// ==================== HTTP сервер ====================
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "Pironman5-Go работает!", "version": "0.1-test"})
	})

	r.POST("/rgb/test", func(c *gin.Context) {
		color := c.Query("color")
		if color == "red" {
			for i := 0; i < 4; i++ {
				strip.Leds(0)[i] = 0xFF0000
			}
		} else if color == "blue" {
			for i := 0; i < 4; i++ {
				strip.Leds(0)[i] = 0x0000FF
			}
		} else {
			for i := 0; i < 4; i++ {
				strip.Leds(0)[i] = 0x00FF00
			}
		}
		strip.Render()
		c.JSON(200, gin.H{"ok": true, "color": color})
	})

	log.Println("Сервер слушает :34001")
	r.Run(":34001")
}
