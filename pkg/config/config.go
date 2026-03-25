package config

type RPIRgbStyle string
type RPIGpioFanMode uint64
type RPIGpioFanLed string

const (
	Solid          RPIRgbStyle = "solid"
	Breathing      RPIRgbStyle = "breathing"
	Flow           RPIRgbStyle = "flow"
	FlowReverse    RPIRgbStyle = "flow_reverse"
	Rainbow        RPIRgbStyle = "rainbow"
	RainbowReverse RPIRgbStyle = "rainbow_reverse"
	HueCycle       RPIRgbStyle = "hue_cycle"
)

const (
	AlwaysOn    RPIGpioFanMode = iota
	Performance RPIGpioFanMode = iota
	Cool        RPIGpioFanMode = iota
	Balance     RPIGpioFanMode = iota
	Silent      RPIGpioFanMode = iota
)

const (
	On     RPIGpioFanLed = "on"
	Off    RPIGpioFanLed = "off"
	Follow RPIGpioFanLed = "follow"
)

type RPIConfigDTO struct {
	RgbColor              string         `json:"rgb_color"`                // hex format (#0a1aff)
	RgbBrightness         uint64         `json:"rgb_brightness"`           // range 0-100
	RgbStyle              RPIRgbStyle    `json:"rgb_style"`                // "solid" | "breathing" | "flow" | "flow_reverse" | "rainbow" | "rainbow_reverse" | "hue_cycle"
	RgbSpeed              uint64         `json:"rgb_speed"`                // range 0-100
	RgbEnabled            bool           `json:"rgb_enabled"`              // true | false
	OledEnabled           bool           `json:"oled_enabled"`             // true | false
	OledDisk              string         `json:"oled_disk"`                // "total" | get_disks()
	OledNetworkInterface  string         `json:"oled_network_interface"`   // "all" | get_ips().keys()
	OledSleepTimeout      uint64         `json:"oled_sleep_timeout"`       // range 0-18446744073709551615
	GpioFanMode           RPIGpioFanMode `json:"gpio_fan_mode"`            // range 0-4
	GpioFanLed            RPIGpioFanLed  `json:"gpio_fan_led"`             // "on" | "off" | "follow"
	VibrationSwitchPullUp bool           `json:"vibration_switch_pull_up"` // true | false
}
