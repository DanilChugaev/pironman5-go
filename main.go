package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/warthog618/go-gpiocdev"
)

const (
	gpioChip = "gpiochip0" // на Pi 5 обычно gpiochip0
	dataPin  = 10          // GPIO10 = BCM 10, физ. пин 19 (MOSI)
	ledCount = 4
	httpPort = ":34001"
)

var leds [ledCount]uint32 // 0xRRGGBB

var line *gpiocdev.Line

func initGPIO() error {
	var err error
	line, err = gpiocdev.RequestLine(gpioChip, dataPin,
		gpiocdev.AsOutput(0), // начальное значение low
		gpiocdev.WithConsumer("ws2812-data"),
	)
	if err != nil {
		return fmt.Errorf("request line failed: %w", err)
	}
	fmt.Printf("✅ GPIO %d открыт (chip %s)\n", dataPin, gpioChip)
	return nil
}

// sendBit отправляет один бит (очень критично к таймингам!)
func sendBit(bit bool) {
	if bit {
		_ = line.SetValue(1)
		time.Sleep(850 * time.Nanosecond) // T1H ~0.7–0.9 мкс
		_ = line.SetValue(0)
		time.Sleep(400 * time.Nanosecond) // T1L ~0.6 мкс
	} else {
		_ = line.SetValue(1)
		time.Sleep(400 * time.Nanosecond) // T0H ~0.35 мкс
		_ = line.SetValue(0)
		time.Sleep(850 * time.Nanosecond) // T0L ~0.8 мкс
	}
}

func sendColor(color uint32) {
	// WS2812 ожидает GRB порядок
	g := byte((color >> 8) & 0xFF)
	r := byte((color >> 16) & 0xFF)
	b := byte(color & 0xFF)

	for _, octet := range []byte{g, r, b} {
		for i := 7; i >= 0; i-- {
			sendBit((octet>>i)&1 == 1)
		}
	}
}

func updateLEDs() {
	if line == nil {
		return
	}

	// Reset >50 мкс
	_ = line.SetValue(0)
	time.Sleep(300 * time.Microsecond)

	for i := 0; i < ledCount; i++ {
		sendColor(leds[i])
	}

	// Финальный reset
	_ = line.SetValue(0)
	time.Sleep(500 * time.Microsecond)
}

func main() {
	runtime.LockOSThread()
	fmt.Println("🚀 Pironman5-Go v0.9 — go-gpiocdev bit-bang WS2812")

	if err := initGPIO(); err != nil {
		log.Fatalf("GPIO init failed: %v\nЗапусти с sudo!", err)
	}
	defer line.Close()

	// Тест
	for i := range leds {
		leds[i] = 0xFF0000 // красный
	}
	updateLEDs()
	time.Sleep(4 * time.Second)
	for i := range leds {
		leds[i] = 0
	}
	updateLEDs()

	// graceful shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		fmt.Println("\nВыключаем...")
		for i := range leds {
			leds[i] = 0
		}
		updateLEDs()
		line.Close()
		os.Exit(0)
	}()

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "driver": "go-gpiocdev bitbang"})
	})

	r.POST("/rgb", func(c *gin.Context) {
		col := c.Query("c")
		var clr uint32
		switch col {
		case "red":
			clr = 0xFF0000
		case "green":
			clr = 0x00FF00
		case "blue":
			clr = 0x0000FF
		case "off":
			clr = 0
		default:
			c.JSON(400, gin.H{"error": "unknown color"})
			return
		}
		for i := range leds {
			leds[i] = clr
		}
		updateLEDs()
		c.JSON(200, gin.H{"ok": true, "color": col})
	})

	r.POST("/rgb/test", func(c *gin.Context) {
		col := c.Query("c")
		var clr uint32
		switch col {
		case "red":
			clr = 0xFF0000
		case "green":
			clr = 0x00FF00
		case "blue":
			clr = 0x0000FF
		case "off":
			clr = 0
		default:
			c.JSON(400, gin.H{"error": "unknown color"})
			return
		}

		line.SetValue(0)
		time.Sleep(1 * time.Millisecond)

		for i := range leds {
			leds[i] = clr
		}
		updateLEDs()
		c.JSON(200, gin.H{"ok": true, "color": col, "value": fmt.Sprintf("0x%06x", clr)})
	})

	log.Printf("Сервер на %s", httpPort)
	r.Run(httpPort)
}
