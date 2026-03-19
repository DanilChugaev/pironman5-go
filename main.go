package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stianeikeland/go-rpio/v4"
)

const ledPin = 10 // GPIO10 — именно то, что использует Pironman5

var leds [4]uint32 // 0xRRGGBB

func sendWS2812(pin rpio.Pin, color uint32) {
	// Простой, но надёжный bit-bang для WS2812 (GRB порядок)
	pin.Output()
	pin.Low()
	time.Sleep(50 * time.Microsecond)

	for i := 0; i < 24; i++ {
		bit := (color >> uint(23-i)) & 1
		if bit == 1 {
			pin.High()
			// T1H ≈ 0.8 мкс
			time.Sleep(800 * time.Nanosecond)
			pin.Low()
			time.Sleep(450 * time.Nanosecond)
		} else {
			pin.High()
			time.Sleep(400 * time.Nanosecond)
			pin.Low()
			time.Sleep(850 * time.Nanosecond)
		}
	}
	pin.Low()
}

func updateLEDs() {
	rpio.Open()
	defer rpio.Close()
	pin := rpio.Pin(ledPin)
	pin.Mode(rpio.Output)

	for i := 0; i < 4; i++ {
		sendWS2812(pin, leds[i])
	}
	// RES (низкий уровень >50 мкс)
	time.Sleep(100 * time.Microsecond)
}

func main() {
	fmt.Println("🚀 Pironman5-Go v0.3 (Pi 5 native, без ws281x)")

	rpio.Open()
	defer rpio.Close()

	// Тест при запуске — все красные 5 сек
	for i := range leds {
		leds[i] = 0xFF0000
	}
	updateLEDs()
	time.Sleep(5 * time.Second)
	for i := range leds {
		leds[i] = 0x000000
	}
	updateLEDs()

	// HTTP
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "OK", "message": "Работает на Pi 5!", "leds": "готовы"})
	})

	r.POST("/rgb", func(c *gin.Context) {
		col := c.Query("c")
		switch col {
		case "red":
			for i := range leds {
				leds[i] = 0xFF0000
			}
		case "green":
			for i := range leds {
				leds[i] = 0x00FF00
			}
		case "blue":
			for i := range leds {
				leds[i] = 0x0000FF
			}
		case "off":
			for i := range leds {
				leds[i] = 0x000000
			}
		}
		updateLEDs()
		c.JSON(200, gin.H{"ok": true, "color": col})
	})

	log.Println("🌐 Сервер запущен → http://0.0.0.0:34001")
	r.Run(":34001")
}
