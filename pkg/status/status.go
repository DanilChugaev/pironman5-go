package status

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

type RPIStatusDTO struct {
	CPUTemperature string `json:"cpu_temperature"`
	// GPUTemperature        float64  `json:"gpu_temperature"`
	// CpuPercent            float64  `json:"cpu_percent"`
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

func runPythonCommand(method string) string {
	script := "scripts/status/" + method + ".py"
	cmd := exec.Command("venv/bin/python3", script)

	// Запуск и получение вывода
	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("Ошибка запуска: %s", err)
	}

	return string(output)
}

func GetStatus() RPIStatusDTO {
	return RPIStatusDTO{
		CPUTemperature: getCpuTemperature(),
	}
}

func PrintStatus() {
	fmt.Println(runPythonCommand("print_status"))
}

func getCpuTemperature() string {
	return strings.ReplaceAll(runPythonCommand("get_cpu_temperature"), "/n", "")
}
