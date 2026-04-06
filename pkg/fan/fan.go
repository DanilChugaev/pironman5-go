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
	GpioFanPin    int = 6
	GpioFanLedPin int = 5
)

const pythonScript = "scripts/rpi_fan/set_fan.py"

// StartFanControlLoop — горутина с тикером
func StartFanControlLoop(fanUpdateInterval uint64) {
	ticker := time.NewTicker(time.Duration(fanUpdateInterval) * time.Second)
	defer ticker.Stop()

	log.Println("🚀 Fan control loop started (official levels + hysteresis)")

	level := 0 // начальный уровень

	for range ticker.C {
		cfg, err := config.LoadConfig()
		if err != nil {
			log.Printf("fan: failed to load config: %v", err)
			continue
		}

		ticker.Reset(time.Duration(cfg.FanUpdateInterval) * time.Second)

		temp := status.GetCpuTemperature()
		fan_levels := cfg.FanLevels

		// === Логика уровней с гистерезисом (как в официальном fan_service.py) ===
		if temp < fan_levels[level].Low {
			level--
		} else if temp > fan_levels[level].High {
			level++
		}
		// Ограничиваем уровень
		if level < 0 {
			level = 0
		}
		if level >= len(fan_levels) {
			level = len(fan_levels) - 1
		}

		// Включаем gpio_fan, если уровень >= gpio_fan_mode
		fanOn := level >= cfg.GpioFanMode

		// === LED логика (точно как в официальном коде) ===
		ledState := 0
		switch cfg.GpioFanLed {
		case "follow":
			if fanOn {
				ledState = 1
			}
		case "on":
			ledState = 1
		case "off":
			ledState = 0
		default:
			ledState = 0 // fallback
		}

		if err := setFanAndLed(GpioFanPin, fanOn, GpioFanLedPin, ledState); err != nil {
			log.Printf("fan+led: set failed: %v", err)
		} else {
			fanStr := "ON"
			if !fanOn {
				fanStr = "OFF"
			}
			ledStr := "ON"
			if ledState == 0 {
				ledStr = "OFF"
			}
			log.Printf("Fan GPIO%d=%s | LED GPIO%d=%s | Temp %.1f°C | Level %d (%s)",
				GpioFanPin, fanStr, GpioFanLedPin, ledStr, temp, level, fan_levels[level].Name)
		}
	}
}

func setFanAndLed(fanPin int, fanOn bool, ledPin int, ledState int) error {
	scriptPath, _ := filepath.Abs(pythonScript)

	fanState := 0
	if fanOn {
		fanState = 1
	}

	cmd := exec.Command("python3", scriptPath,
		fmt.Sprintf("%d", fanPin), fmt.Sprintf("%d", fanState),
		fmt.Sprintf("%d", ledPin), fmt.Sprintf("%d", ledState),
	)

	output, err := cmd.CombinedOutput()
	log.Printf("Fan+LED python output: %s", strings.TrimSpace(string(output)))

	if err != nil {
		return fmt.Errorf("python error: %v | output: %s", err, output)
	}
	return nil
}
