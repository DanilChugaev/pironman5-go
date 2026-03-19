package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stianeikeland/go-rpio/v4"
)

// Конфигурация
const (
	ledPin   = 10 // GPIO10 (BCM numbering) для Pironman5
	httpPort = ":34001"
	ledCount = 4
)

// Глобальное состояние
var (
	pin      rpio.Pin
	leds     [ledCount]uint32 // Хранение цвета в формате 0xRRGGBB
	gpioOpen bool
)

// initGPIO инициализирует GPIO один раз при старте
func initGPIO() error {
	// Попытка открыть GPIO
	// На Pi 5 иногда требуется явно указать версию, но обычно библиотека сама определяет.
	// Если возникает ошибка здесь, проверьте права доступа (sudo) и версию библиотеки.
	if err := rpio.Open(); err != nil {
		return fmt.Errorf("failed to open gpio: %w", err)
	}

	pin = rpio.Pin(ledPin)
	pin.Mode(rpio.Output)
	pin.Low()

	gpioOpen = true
	fmt.Println("✅ GPIO initialized successfully")
	return nil
}

// closeGPIO безопасно закрывает соединение
func closeGPIO() {
	if gpioOpen {
		rpio.Close()
		gpioOpen = false
		fmt.Println("🛑 GPIO closed")
	}
}

// sendWS2812 отправляет данные на один светодиод
// Внимание: time.Sleep в Linux не гарантирует точность до наносекунд.
// Для продакшена на Pi 5 настоятельно рекомендуется использовать библиотеку,
// работающую через демон pigpio (например, github.com/rpi-ws281x-go/ws281x),
// но этот код оставляет битбанг для зависимости только от stdlib + rpio.
func sendWS2812(color uint32) {
	if !gpioOpen {
		log.Println("⚠️ GPIO not open, skipping LED update")
		return
	}

	// WS2812 использует порядок байт GRB, но мы храним как RGB.
	// Нужно пересобрать биты или менять логику формирования цвета.
	// В исходном коде автора цвет передавался как есть, предположим, что пользователь передает уже готовый паттерн
	// или библиотека/светодиоды принимают RGB.
	// Стандарт WS2812: Green first, then Red, then Blue.
	// Если у вас цвета смешиваются, нужно сделать своп байтов здесь.

	// Преобразуем 0xRRGGBB в поток битов для отправки (предполагаем, что входной цвет уже в нужном порядке или светодиоды простые)
	// Для классических WS2812B порядок данных: G7..G0, R7..R0, B7..B0

	grbColor := ((color & 0x00FF00) << 8) | ((color & 0xFF0000) >> 8) | (color & 0x0000FF)

	for i := 23; i >= 0; i-- {
		bit := (grbColor >> uint(i)) & 1
		if bit == 1 {
			pin.High()
			time.Sleep(800 * time.Nanosecond) // T1H
			pin.Low()
			time.Sleep(450 * time.Nanosecond) // T1L
		} else {
			pin.High()
			time.Sleep(400 * time.Nanosecond) // T0H
			pin.Low()
			time.Sleep(850 * time.Nanosecond) // T0L
		}
	}
}

// updateLEDs обновляет всю ленту
func updateLEDs() {
	if !gpioOpen {
		return
	}

	// Сброс линии перед передачей (не всегда обязательно, но полезно)
	pin.Low()
	time.Sleep(60 * time.Microsecond)

	for i := 0; i < ledCount; i++ {
		sendWS2812(leds[i])
	}

	// Reset signal (низкий уровень > 50 мкс)
	pin.Low()
	time.Sleep(100 * time.Microsecond)
}

// setAllColors устанавливает цвет всем светодиодам
func setAllColors(color uint32) {
	for i := range leds {
		leds[i] = color
	}
	updateLEDs()
}

func main() {
	fmt.Printf("🚀 Pironman5-Go v0.4\n")

	// Инициализация GPIO
	if err := initGPIO(); err != nil {
		log.Fatalf("❌ Critical error initializing GPIO: %v\nЗапустите программу с sudo или проверьте права на /dev/gpiomem", err)
	}

	// Обработка сигналов завершения для корректного закрытия GPIO
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n🛑 Shutting down...")
		setAllColors(0x000000) // Выключаем свет перед выходом
		closeGPIO()
		os.Exit(0)
	}()

	// Тестовый прогон при старте
	fmt.Println("💡 Running startup test (Red -> Off)")
	setAllColors(0xFF0000) // Красный
	time.Sleep(1 * time.Second)
	setAllColors(0x000000) // Выкл

	// Настройка Gin
	gin.SetMode(gin.ReleaseMode) // Тише логи в консоли
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":     "OK",
			"version":    "v0.4-refactor",
			"platform":   "Raspberry Pi 5",
			"leds_count": ledCount,
		})
	})

	r.POST("/rgb", func(c *gin.Context) {
		col := c.Query("c")
		var targetColor uint32
		var name string

		switch col {
		case "red":
			targetColor = 0xFF0000
			name = "red"
		case "green":
			targetColor = 0x00FF00
			name = "green"
		case "blue":
			targetColor = 0x0000FF
			name = "blue"
		case "white":
			targetColor = 0xFFFFFF
			name = "white"
		case "off":
			targetColor = 0x000000
			name = "off"
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid color. use: red, green, blue, white, off"})
			return
		}

		setAllColors(targetColor)
		c.JSON(http.StatusOK, gin.H{"ok": true, "color": name, "hex": fmt.Sprintf("%06X", targetColor)})
	})

	// Запуск сервера
	log.Printf("🌐 Server started on http://0.0.0.0%s", httpPort)
	if err := r.Run(httpPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
