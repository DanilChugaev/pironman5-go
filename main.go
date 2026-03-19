package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	ws2811 "github.com/rpi-ws281x/rpi-ws281x-go"
)

func main() {
	fmt.Println("🚀 Pironman5-Go v0.2 запущен")

	// ==================== WS281x ====================
	opt := &ws2811.Option{
		Frequency: ws2811.TargetFreq, // 800 кГц
		DmaNum:    ws2811.DefaultDmaNum,
		Channels: []ws2811.ChannelOption{
			{
				GpioPin:    10, // GPIO 10 = SPI MOSI — идеально для Pironman5 + Pi 5
				LedCount:   4,
				Brightness: 64,                 // 0-255, можно менять потом
				StripeType: ws2811.WS2812Strip, // GRB порядок (стандарт для WS2812)
				Invert:     false,
			},
		},
	}

	dev, err := ws2811.MakeWS2811(opt)
	if err != nil {
		log.Fatalf("MakeWS2811 ошибка: %v", err)
	}
	defer dev.Fini() // ← правильный способ очистки

	if err := dev.Init(); err != nil {
		log.Fatalf("Init ошибка: %v", err)
	}

	// Тест при запуске: все светодиоды красные 5 сек
	leds := dev.Leds(0)
	for i := range leds {
		leds[i] = 0xFF0000 // красный (работает с WS2812)
	}
	if err := dev.Render(); err != nil {
		log.Println("Render:", err)
	}
	time.Sleep(5 * time.Second)

	// Выключаем
	for i := range leds {
		leds[i] = 0x000000
	}
	dev.Render()

	// ==================== HTTP ====================
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "OK", "leds": "готовы", "gpio": 10})
	})

	r.POST("/rgb", func(c *gin.Context) {
		color := c.Query("c") // ?c=red / blue / green / off
		switch color {
		case "red":
			for i := range leds {
				leds[i] = 0xFF0000
			}
		case "blue":
			for i := range leds {
				leds[i] = 0x0000FF
			}
		case "green":
			for i := range leds {
				leds[i] = 0x00FF00
			}
		default:
			for i := range leds {
				leds[i] = 0x000000
			}
		}
		dev.Render()
		c.JSON(200, gin.H{"ok": true, "color": color})
	})

	log.Println("🌐 Сервер на http://0.0.0.0:34001")
	r.Run(":34001")
}
