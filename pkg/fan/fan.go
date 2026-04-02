package fan

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/DanilChugaev/pironman5-go/pkg/config"
	"github.com/DanilChugaev/pironman5-go/pkg/status"
)

const (
	GpioFanPin int = 6
)

const pythonScript = "scripts/rpi_fan/set_fan_pwm.py"

// StartFanControlLoop запускает горутину с тикером
func StartFanControlLoop(fanUpdateInterval uint64) {
	ticker := time.NewTicker(time.Duration(fanUpdateInterval) * time.Second) // по умолчанию, будет браться из конфига
	defer ticker.Stop()

	log.Println("Система управления вентиляторами запущена! (PWM)")

	for range ticker.C {
		// Перезагружаем конфиг каждый тик — чтобы изменения через PUT /api/config сразу применялись
		cfg, err := config.LoadConfig()
		if err != nil {
			log.Printf("fan: failed to load config: %v", err)
			continue
		}

		// Обновляем интервал тикера "на лету"
		ticker.Reset(time.Duration(cfg.FanUpdateInterval) * time.Second)

		temp := status.GetCpuTemperature()

		mainThreshold := cfg.FanMainStartTemp
		addThreshold := cfg.FanAddStartTemp

		var duty float64
		switch {
		case temp >= addThreshold:
			duty = 100.0 // полная мощность — оба дополнительных вентилятора
		case temp >= mainThreshold:
			duty = 50.0 // частичная мощность — "основной режим"
		default:
			duty = 0.0 // выключено
		}

		// Вызываем Python (PWM)
		if err := setFanPWM(GpioFanPin, duty); err != nil {
			log.Printf("fan: set PWM failed: %v", err)
		} else {
			log.Printf("Fan GPIO%d | Temp %.1f°C | Duty %.0f%% (main:%.0f add:%.0f)", GpioFanPin, temp, duty, mainThreshold, addThreshold)
		}
	}
}

// setFanPWM вызывает Python-скрипт
func setFanPWM(pin int, dutyPercent float64) error {
	scriptPath, err := filepath.Abs(pythonScript)
	if err != nil {
		return err
	}

	cmd := exec.Command("python3", scriptPath, fmt.Sprintf("%d", pin), fmt.Sprintf("%.0f", dutyPercent))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("python error: %v | output: %s", err, output)
	}
	return nil
}
