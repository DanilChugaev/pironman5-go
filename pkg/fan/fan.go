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
	FanGpioPin    int = 6
	FanGpioLedPin int = 5
)

const (
	pythonFanScript   = "scripts/rpi_fan/set_fan.py"
	pythonTowerScript = "scripts/rpi_fan/set_tower_fan.py"
	towerHysteresis   = 5.0 // °C гистерезис для tower-фана
)

// StartFanControlLoop — горутина с тикером
func StartFanControlLoop(fanUpdateInterval uint64) {
	ticker := time.NewTicker(time.Duration(fanUpdateInterval) * time.Second)
	defer ticker.Stop()

	log.Println("🚀 Fan + LED + Tower PWM control loop started")

	level := 0    // уровень для GPIO-вентиляторов
	towerPWM := 0 // 0-4 для tower-фана

	for range ticker.C {
		cfg, err := config.LoadConfig()
		if err != nil {
			log.Printf("fan: failed to load config: %v", err)
			continue
		}

		ticker.Reset(time.Duration(cfg.FanUpdateInterval) * time.Second)

		temp := status.GetCpuTemperature()
		fan_levels := cfg.FanLevels

		// === 1. GPIO-вентиляторы ===
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
		fanOn := level >= cfg.FanGpioMode

		// === 2. LED вентиляторов ===
		ledState := 0
		switch cfg.FanGpioLed {
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

		// === 3. Tower-фан (PWM) с отдельной температурой и гистерезисом ===
		startTemp := cfg.FanTowerStartTemp
		if temp >= startTemp {
			// Включаем и повышаем скорость в зависимости от температуры
			if temp >= startTemp+15 {
				towerPWM = 4
			} else if temp >= startTemp+10 {
				towerPWM = 3
			} else if temp >= startTemp+5 {
				towerPWM = 2
			} else {
				towerPWM = 1 // минимальная скорость при старте
			}
		} else if temp < startTemp-towerHysteresis {
			towerPWM = 0 // выключаем только с гистерезисом
		}
		// (уровень towerPWM сохраняется между тиками — это и есть гистерезис)

		// Применяем всё
		if err := setFanAndLed(FanGpioPin, fanOn, FanGpioLedPin, ledState); err != nil {
			log.Printf("fan+led: %v", err)
		}
		if err := setTowerFan(towerPWM); err != nil {
			log.Printf("tower: %v", err)
		} else {
			fanStr := map[bool]string{true: "ON", false: "OFF"}[fanOn]
			ledStr := map[int]string{1: "ON", 0: "OFF"}[ledState]
			log.Printf("GPIO Fan=%s | LED=%s | Tower PWM=%d | Temp %.1f°C | Level %d (%s) | TowerStart %.0f°C",
				fanStr, ledStr, towerPWM, temp, level, fan_levels[level].Name, startTemp)
		}
	}
}

// === fan + led ===
func setFanAndLed(fanPin int, fanOn bool, ledPin int, ledState int) error {
	scriptPath, _ := filepath.Abs(pythonFanScript)

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

// === tower-fan ===
func setTowerFan(pwm int) error {
	scriptPath, _ := filepath.Abs(pythonTowerScript)
	cmd := exec.Command("python3", scriptPath, fmt.Sprintf("%d", pwm))
	output, err := cmd.CombinedOutput()
	log.Printf("Tower python output: %s", strings.TrimSpace(string(output)))
	if err != nil {
		return fmt.Errorf("python error: %v | output: %s", err, output)
	}
	return nil
}
