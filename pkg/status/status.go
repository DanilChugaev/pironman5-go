package status

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

type RPIStatusDTO struct {
	CPUTemperature float64 `json:"cpu_temperature"`
	GPUTemperature float64 `json:"gpu_temperature"`
	CpuPercent     float64 `json:"cpu_percent"`
	// CpuPercentPerCPU      any      `json:"cpu_percent_per_cpu"`
	// CpuFrequency          string   `json:"cpu_frequency"`
	// CpuCount              uint     `json:"cpu_count"`
	// MemoryInfo            string   `json:"memory_info"`
	// DiskInfo              string   `json:"disk_info"`
	// DiskInfoPerDisk       any      `json:"disk_info_per_disk"`
	// Disks                 []string `json:"disks"`
	// BootTime              float64  `json:"boot_time"`
	// Ips                   any      `json:"ips"`
	// Macs                  any      `json:"macs"`
	// NetworkConnectionType any      `json:"network_connection_type"`
	// NetworkSpeed          any      `json:"network_speed"`
}

// == внешние методы ==

func GetStatus() RPIStatusDTO {
	return RPIStatusDTO{
		CPUTemperature: getCpuTemperature(),
		GPUTemperature: getGpuTemperature(),
		CpuPercent:     getCpuPercent(),
	}
}

func PrintStatus() {
	fmt.Println(runPythonCommand("print_status"))
}

// == внутренние методы ==

func runPythonCommand(method string) string {
	script := "scripts.rpi_status.methods." + method
	cmd := exec.Command("venv/bin/python3", "-m", script)

	// Запуск и получение вывода
	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("Ошибка запуска: %s", err)
	}

	return string(output)
}

func strToFloat(str string) float64 {
	result, err := strconv.ParseFloat(strings.ReplaceAll(str, "\n", ""), 64)

	if err != nil {
		fmt.Println("Ошибка преобразования:", err)
		return 0.0
	}

	return result
}

// == todo: заменить реализацию методов на чистый GO ==

func getCpuTemperature() float64 {
	return strToFloat(runPythonCommand("get_cpu_temperature"))
}

func getGpuTemperature() float64 {
	return strToFloat(runPythonCommand("get_gpu_temperature"))
}

func getCpuPercent() float64 {
	return strToFloat(runPythonCommand("get_cpu_percent"))
}
