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

// FAN_LEVELS — как в официальном SunFounder
var FAN_LEVELS = []struct {
	Name    string
	Low     float64
	High    float64
	Percent int
}{
	{"OFF", -200, 45, 0},
	{"LOW", 35, 55, 40},
	{"MEDIUM", 45, 65, 60},
	{"HIGH", 55, 100, 100},
}

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

		// === Логика уровней с гистерезисом (как в официальном fan_service.py) ===
		changed := false
		if temp < FAN_LEVELS[level].Low {
			level--
			changed = true
		} else if temp > FAN_LEVELS[level].High {
			level++
			changed = true
		}

		// Ограничиваем уровень
		if level < 0 {
			level = 0
		}
		if level >= len(FAN_LEVELS) {
			level = len(FAN_LEVELS) - 1
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
					GpioFanPin, temp, statusStr, level, FAN_LEVELS[level].Name, FAN_LEVELS[level].Percent)
			} else {
				log.Printf("Fan GPIO%d | Temp %.1f°C | %s (level %d: %s)",
					GpioFanPin, temp, statusStr, level, FAN_LEVELS[level].Name)
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
