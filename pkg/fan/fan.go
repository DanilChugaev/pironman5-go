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
	Gist       int = 5
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
		changed := false
		if temp < fan_levels[level].High-float64(Gist) {
			level--
			changed = true
		} else if temp > fan_levels[level].High {
			level++
			changed = true
		}

		// Ограничиваем уровень
		if level < 0 {
			level = 0
		}
		if level >= len(fan_levels) {
			level = len(fan_levels) - 1
		}

		// Включаем gpio_fan, если уровень >= gpio_fan_mode
		on := level >= cfg.GpioFanMode

		if err := setFan(GpioFanPin, on); err != nil {
			log.Printf("fan: set failed: %v", err)
		} else {
			statusStr := "ON"
			if !on {
				statusStr = "OFF"
			}
			if changed {
				log.Printf("Fan GPIO%d | Temp %.1f°C → %s (level %d: %s, power %d%%)",
					GpioFanPin, temp, statusStr, level, fan_levels[level].Name, fan_levels[level].High)
			} else {
				log.Printf("Fan GPIO%d | Temp %.1f°C | %s (level %d: %s)",
					GpioFanPin, temp, statusStr, level, fan_levels[level].Name)
			}
		}
	}
}

func setFan(pin int, on bool) error {
	scriptPath, _ := filepath.Abs(pythonScript)

	state := 0
	if on {
		state = 1
	}

	cmd := exec.Command("python3", scriptPath, fmt.Sprintf("%d", pin), fmt.Sprintf("%d", state))
	output, err := cmd.CombinedOutput()

	log.Printf("Fan python output: %s", strings.TrimSpace(string(output)))

	if err != nil {
		return fmt.Errorf("python error: %v | output: %s", err, output)
	}
	return nil
}
