package fan

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/DanilChugaev/pironman5-go/pkg/config"
	"github.com/DanilChugaev/pironman5-go/pkg/status"
)

const (
	GpioFanPin int = 6
)

const pythonScript = "scripts/rpi_fan/set_fan.py"

// StartFanControlLoop — горутина с тикером
func StartFanControlLoop(fanUpdateInterval uint64) {
	ticker := time.NewTicker(time.Duration(fanUpdateInterval) * time.Second)
	defer ticker.Stop()

	log.Println("🚀 Fan control loop started (on/off via GPIO6)")

	for range ticker.C {
		cfg, err := config.LoadConfig()
		if err != nil {
			log.Printf("fan: failed to load config: %v", err)
			continue
		}

		// Обновляем интервал на лету
		ticker.Reset(time.Duration(cfg.FanUpdateInterval) * time.Second)

		temp := status.GetStatus().CPUTemperature

		mainThreshold := cfg.FanMainStartTemp
		addThreshold := cfg.FanAddStartTemp

		// Простая логика (можно потом добавить hysteresis)
		on := false
		if temp >= addThreshold || temp >= mainThreshold {
			on = true
		}

		if err := setFan(GpioFanPin, on); err != nil {
			log.Printf("fan: set failed: %v", err)
		} else {
			statusStr := "ON"
			if !on {
				statusStr = "OFF"
			}
			log.Printf("Fan GPIO%d | Temp %.1f°C | %s (main:%.0f add:%.0f)", GpioFanPin, temp, statusStr, mainThreshold, addThreshold)
		}
	}
}

func setFan(pin int, on bool) error {
	scriptPath, _ := filepath.Abs(pythonScript) // если не сработает — замени на полный путь

	state := 0
	if on {
		state = 1
	}

	cmd := exec.Command("python3", scriptPath, fmt.Sprintf("%d", pin), fmt.Sprintf("%d", state))
	output, err := cmd.CombinedOutput()

	// ВАЖНО: всегда логируем вывод Python, даже при успехе
	log.Printf("Fan python output: %s", strings.TrimSpace(string(output)))

	if err != nil {
		return fmt.Errorf("python error: %v | output: %s", err, output)
	}
	return nil
}
