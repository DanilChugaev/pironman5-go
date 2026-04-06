package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type FanLevel struct {
	Name string  `json:"name"`
	Low  float64 `json:"low"`
	High float64 `json:"high"`
}

const (
	Solid          string = "solid"
	Breathing      string = "breathing"
	Flow           string = "flow"
	FlowReverse    string = "flow_reverse"
	Rainbow        string = "rainbow"
	RainbowReverse string = "rainbow_reverse"
	HueCycle       string = "hue_cycle"
)

const (
	AlwaysOn    int = iota
	Performance int = iota
	Cool        int = iota
	Balance     int = iota
	Silent      int = iota
)

const (
	On     string = "on"
	Off    string = "off"
	Follow string = "follow"
)

type RPIConfigDTO struct {
	RgbColor              string     `json:"rgb_color"`                // hex format (#0a1aff)
	RgbBrightness         uint64     `json:"rgb_brightness"`           // range 0-100
	RgbStyle              string     `json:"rgb_style"`                // "solid" | "breathing" | "flow" | "flow_reverse" | "rainbow" | "rainbow_reverse" | "hue_cycle"
	RgbSpeed              uint64     `json:"rgb_speed"`                // range 0-100
	RgbEnabled            bool       `json:"rgb_enabled"`              // true | false
	OledEnabled           bool       `json:"oled_enabled"`             // true | false
	OledDisk              string     `json:"oled_disk"`                // "total" | get_disks()
	OledNetworkInterface  string     `json:"oled_network_interface"`   // "all" | get_ips().keys()
	OledSleepTimeout      uint64     `json:"oled_sleep_timeout"`       // range 0-18446744073709551615
	VibrationSwitchPullUp bool       `json:"vibration_switch_pull_up"` // true | false
	FanGpioMode           int        `json:"fan_gpio_mode"`            // range 0-4
	FanGpioLed            string     `json:"fan_gpio_led"`             // "on" | "off" | "follow"
	FanUpdateInterval     uint64     `json:"fan_update_interval"`      // секунды, default 5
	FanLevels             []FanLevel `json:"fan_levels"`               // уровни работы fan вентиляторов - "OFF" | "LOW" | "MEDIUM" | "HIGH"
	FanTowerStartTemp     float64    `json:"fan_tower_start_temp"`     // °C, по умолчанию 50.0
}

// --- Структура для частичного обновления ---
// Поля являются указателями
// Если поле nil, оно не обновляется
type RPIConfigUpdate struct {
	RgbColor              *string     `json:"rgb_color,omitempty"`
	RgbBrightness         *uint64     `json:"rgb_brightness,omitempty"`
	RgbStyle              *string     `json:"rgb_style,omitempty"`
	RgbSpeed              *uint64     `json:"rgb_speed,omitempty"`
	RgbEnabled            *bool       `json:"rgb_enabled,omitempty"`
	OledEnabled           *bool       `json:"oled_enabled,omitempty"`
	OledDisk              *string     `json:"oled_disk,omitempty"`
	OledNetworkInterface  *string     `json:"oled_network_interface,omitempty"`
	OledSleepTimeout      *uint64     `json:"oled_sleep_timeout,omitempty"`
	VibrationSwitchPullUp *bool       `json:"vibration_switch_pull_up,omitempty"`
	FanGpioMode           *int        `json:"fan_gpio_mode,omitempty"`
	FanGpioLed            *string     `json:"fan_gpio_led,omitempty"`
	FanUpdateInterval     *uint64     `json:"fan_update_interval,omitempty"`
	FanLevels             *[]FanLevel `json:"fan_levels,omitempty"`
	FanTowerStartTemp     *float64    `json:"fan_tower_start_temp,omitempty"`
}

const CONFIG_PATH = "pkg/config/config.json"

// == вспомогательные функции ==

// getDefaultValue возвращает конфигурацию с дефолтными настройками
func getDefaultValue() RPIConfigDTO {
	return RPIConfigDTO{
		RgbColor:              "#0a1aff",
		RgbBrightness:         50,
		RgbStyle:              Breathing,
		RgbSpeed:              50,
		RgbEnabled:            true,
		OledEnabled:           true,
		OledDisk:              "total",
		OledNetworkInterface:  "all",
		OledSleepTimeout:      10,
		FanGpioMode:           AlwaysOn,
		FanGpioLed:            Follow,
		VibrationSwitchPullUp: false,
		FanUpdateInterval:     5,
		FanLevels: []FanLevel{
			{"OFF", -200, 55.0},
			{"LOW", 45.0, 65.0},
			{"MEDIUM", 55.0, 75.0},
			{"HIGH", 65.0, 100.0},
		},
		FanTowerStartTemp: 50.0,
	}
}

// writeConfigFile записывает структуру в файл
func writeConfigFile(path string, cfg *RPIConfigDTO) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(path)

	// Права 0644: чтение/запись для владельца, чтение для остальных
	return os.WriteFile(path, data, 0644)
}

// == чтение или создание конфига ==

// LoadConfig читает файл конфигурации
// Если файла не существует, он создается с дефолтными значениями
// Возвращает указатель на структуру конфигурации
func LoadConfig() (*RPIConfigDTO, error) {
	// Проверяем существование файла
	if _, err := os.Stat(CONFIG_PATH); os.IsNotExist(err) {
		// Файла нет, создаем с дефолтными значениями
		defaultCfg := getDefaultValue()
		if err := writeConfigFile(CONFIG_PATH, &defaultCfg); err != nil {
			return nil, fmt.Errorf("Ошибка создания конфига с дефолтными настройками: %w", err)
		}

		fmt.Println("Конфиг создан с дефолтными настройками")

		return &defaultCfg, nil
	} else if err != nil {
		return nil, fmt.Errorf("Ошибка проверки конфига: %w", err)
	}

	// Файл существует, читаем его
	data, err := os.ReadFile(CONFIG_PATH)
	if err != nil {
		return nil, fmt.Errorf("Ошибка чтения конфига: %w", err)
	}

	var cfg RPIConfigDTO
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("Ошибка парсинга JSON из конфига: %w", err)
	}

	return &cfg, nil
}

// == обновление конфига ==

// UpdateConfig загружает текущий конфиг, обновляет его переданными данными
// и записывает результат обратно в файл.
// Позволяет передавать только часть параметров (через структуру RPIConfigUpdate).
func UpdateConfig(updates *RPIConfigUpdate) (*RPIConfigDTO, error) {
	// 1. Используем предыдущую функцию для получения актуального конфига
	currentCfg, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config for update: %w", err)
	}

	// 2. Обновляем поля только если они не nil (частичное обновление)

	// == адресные rgb светодиоды ==
	if updates.RgbColor != nil {
		currentCfg.RgbColor = *updates.RgbColor
	}
	if updates.RgbBrightness != nil {
		currentCfg.RgbBrightness = *updates.RgbBrightness
	}
	if updates.RgbStyle != nil {
		currentCfg.RgbStyle = *updates.RgbStyle
	}
	if updates.RgbSpeed != nil {
		currentCfg.RgbSpeed = *updates.RgbSpeed
	}
	if updates.RgbEnabled != nil {
		currentCfg.RgbEnabled = *updates.RgbEnabled
	}

	// == oled экран ==
	if updates.OledEnabled != nil {
		currentCfg.OledEnabled = *updates.OledEnabled
	}
	if updates.OledDisk != nil {
		currentCfg.OledDisk = *updates.OledDisk
	}
	if updates.OledNetworkInterface != nil {
		currentCfg.OledNetworkInterface = *updates.OledNetworkInterface
	}
	if updates.OledSleepTimeout != nil {
		currentCfg.OledSleepTimeout = *updates.OledSleepTimeout
	}
	if updates.VibrationSwitchPullUp != nil {
		currentCfg.VibrationSwitchPullUp = *updates.VibrationSwitchPullUp
	}

	// == дополнительные вентиляторы fan ==
	if updates.FanGpioMode != nil {
		currentCfg.FanGpioMode = *updates.FanGpioMode
	}
	if updates.FanGpioLed != nil {
		currentCfg.FanGpioLed = *updates.FanGpioLed
	}
	if updates.FanUpdateInterval != nil {
		currentCfg.FanUpdateInterval = *updates.FanUpdateInterval
	}
	if updates.FanLevels != nil {
		currentCfg.FanLevels = *updates.FanLevels
	}

	// 3. Записываем обновленный конфиг в файл
	if err := writeConfigFile(CONFIG_PATH, currentCfg); err != nil {
		return nil, fmt.Errorf("Ошибка обновления конфига: %w", err)
	}

	return currentCfg, nil
}
