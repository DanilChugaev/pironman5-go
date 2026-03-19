package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"unsafe"

	"github.com/gin-gonic/gin"
)

const (
	spiDevice  = "/dev/spidev0.0" // проверь ls /dev/spidev* после raspi-config → SPI Yes
	ledCount   = 4
	httpPort   = ":34001"
	spiSpeedHz = 3200000 // 3.2 MHz — часто оптимально; можно 2500000 / 4000000
)

var leds [ledCount]uint32 // 0xRRGGBB

// ioctl константы для SPI (из linux/spi/spidev.h)
const (
	SPI_IOC_MESSAGE_BASE = 0x40006b00 // _IOW('k', 0, ...)
	SPI_IOC_WR_MAX_SPEED = 0x40046b04 // _IOW('k', 4, __u32)
)

type spiIocTransfer struct {
	TxBuf       uint64
	RxBuf       uint64
	Len         uint32
	SpeedHz     uint32
	DelayUsecs  uint16
	BitsPerWord uint8
	CSChange    uint8
	TxNbits     uint8
	RxNbits     uint8
	Pad         uint16
}

func setSPISpeed(fd uintptr, speed uint32) error {
	return ioctl(fd, SPI_IOC_WR_MAX_SPEED, uintptr(unsafe.Pointer(&speed)))
}

func ioctl(fd, request, arg uintptr) error {
	_, _, e1 := syscall.Syscall(syscall.SYS_IOCTL, fd, request, arg)
	if e1 != 0 {
		return syscall.Errno(e1)
	}
	return nil
}

func spiTransfer(fd uintptr, tx []byte) error {
	xfers := []spiIocTransfer{
		{
			TxBuf:       uint64(uintptr(unsafe.Pointer(&tx[0]))),
			Len:         uint32(len(tx)),
			SpeedHz:     spiSpeedHz,
			BitsPerWord: 8,
		},
	}

	return ioctl(fd, SPI_IOC_MESSAGE_BASE|uintptr(len(xfers)), uintptr(unsafe.Pointer(&xfers[0])))
}

func rgbToSPIBytes() []byte {
	buf := make([]byte, 0, ledCount*24*3/8+50)

	for _, c := range leds {
		// RGB → GRB
		g := byte((c >> 8) & 0xFF)
		r := byte((c >> 16) & 0xFF)
		b := byte(c & 0xFF)

		for _, octet := range []byte{g, r, b} {
			for bit := 7; bit >= 0; bit-- {
				if (octet>>uint(bit))&1 == 1 {
					buf = append(buf, 0b11000000>>2, 0b00000000, 0b00000000) // подгонка под 3 бита на байт
				} else {
					buf = append(buf, 0b10000000>>2, 0b00000000, 0b00000000)
				}
			}
		}
	}

	// Reset >50us — просто длинная пауза низкого уровня (много нулей)
	for i := 0; i < 50; i++ {
		buf = append(buf, 0)
	}

	return buf
}

func updateLEDs() error {
	f, err := os.OpenFile(spiDevice, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer f.Close()

	fd := f.Fd()

	// Устанавливаем скорость
	if err := setSPISpeed(fd, spiSpeedHz); err != nil {
		return err
	}

	data := rgbToSPIBytes()
	return spiTransfer(fd, data)
}

func main() {
	fmt.Println("🚀 Pironman5-Go v0.7 — чистый SPI + ioctl (без внешних пакетов)")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sig
		fmt.Println("\nВыключаем...")
		for i := range leds {
			leds[i] = 0
		}
		updateLEDs()
		os.Exit(0)
	}()

	// Тест
	for i := range leds {
		leds[i] = 0xFF0000
	}
	if err := updateLEDs(); err != nil {
		log.Fatalf("Ошибка запуска: %v", err)
	}
	time.Sleep(3 * time.Second)
	for i := range leds {
		leds[i] = 0
	}
	updateLEDs()

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "driver": "native SPI ioctl"})
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
			c.JSON(400, gin.H{"error": "неизвестный цвет"})
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

	log.Printf("Сервер запущен на %s", httpPort)
	r.Run(httpPort)
}
