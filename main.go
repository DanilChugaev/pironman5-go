package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/exp/io/spi"
)

const (
	spiDevice   = "/dev/spidev0.0" // или spidev10.0 в зависимости от конфига
	ledCount    = 4
	httpPort    = ":34001"
	spiSpeedHz  = 3200000 // 3.2–4 MHz обычно оптимально для WS2812
)

var leds [ledCount]uint32 // 0xRRGGBB

func rgbToSPIBytes(colors []uint32) []byte {
	// WS2812: 24 бита GRB на LED → 3 байта
	// SPI bit-stuffing: каждый бит → 3 SPI-бита (110 = 1, 100 = 0)
	// → 24 бита → 72 SPI-бита → 9 байт на LED
	buf := make([]byte, ledCount*24*3/8) // 9 байт × ledCount

	pos := 0
	for _, c := range colors {
		// RGB → GRB
		g := byte((c >> 8) & 0xFF)
		r := byte((c >> 16) & 0xFF)
		b := byte(c & 0xFF)

		for _, octet := range []byte{g, r, b} {
			for bit := 7; bit >= 0; bit-- {
				b := (octet >> uint(bit)) & 1
				var pattern byte
				if b == 1 {
					pattern = 0b110 // T1H ≈ 0.8 μs, T1L ≈ 0.45 μs при 3.2 MHz
				} else {
					pattern = 0b100 // T0H ≈ 0.4 μs, T0L ≈ 0.85 μs
				}

				// Распределяем 3 бита по байтам
				buf[pos/8] |= pattern << (5 - (pos % 8)) // старшие биты
				if pos%8 >= 5 {
					buf[(pos/8)+1] |= pattern >> (8 - (5 - (pos % 8)))
				}
				pos += 3
			}
		}
	}

	// Reset > 50 μs → просто добавить нули в конце (SPI сам даст паузу)
	buf = append(buf, make([]byte, 50)...)

	return buf
}

func updateLEDs() error {
	dev, err := spi.Open(&spi.Devfs{
		Dev:      spiDevice,
		Mode:     spi.Mode0,
		MaxSpeed: spiSpeedHz,
	})
	if err != nil {
		return err
	}
	defer dev.Close()

	data := rgbToSPIBytes(leds[:])
	_, err = dev.Write(data)
	return err
}

func main() {
	fmt.Println("🚀 Pironman5-Go v0.6 — SPI bit-stuffing mode")

	// Graceful shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		fmt.Println("\nВыключаем LED...")
		for i := range leds {
			leds[i] = 0
		}
		updateLEDs()
		os.Exit(0)
	}()

	// Тест при запуске
	for i := range leds {
		leds[i] = 0xFF0000 // красный
	}
	if err := updateLEDs(); err != nil {
		log.Fatalf("SPI init failed: %v", err)
	}
	time.Sleep(3 * time.Second)
	for i := range leds {
		leds[i] = 0
	}
	updateLEDs()

	// HTTP
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "driver": "SPI bitbang"})
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
		if err := updateLEDs(); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"ok": true, "color": col})
	})

	log.Printf("Сервер на %s", httpPort)
	r.Run(httpPort)
}