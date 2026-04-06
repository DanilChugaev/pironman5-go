package status

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type RPIStatusDTO struct {
	CPUTemperature        float64 `json:"cpu_temperature"`
	CpuPercent            float64 `json:"cpu_percent"`
	CpuPercentPerCPU      any     `json:"cpu_percent_per_cpu"`
	CpuFrequency          string  `json:"cpu_frequency"`
	CpuCount              uint64  `json:"cpu_count"`
	MemoryInfo            string  `json:"memory_info"`
	DiskInfo              string  `json:"disk_info"`
	Disks                 any     `json:"disks"`
	DiskInfoPerDisk       any     `json:"disk_info_per_disk"`
	BootTime              string  `json:"boot_time"`
	Ips                   any     `json:"ips"`
	Macs                  any     `json:"macs"`
	NetworkConnectionType any     `json:"network_connection_type"`
	NetworkSpeed          any     `json:"network_speed"`
}

// == внешние методы ==

func GetStatus() RPIStatusDTO {
	return RPIStatusDTO{
		CPUTemperature:        GetCpuTemperature(),
		CpuPercent:            getCpuPercent(),
		CpuPercentPerCPU:      getCpuPercentPerCpu(),
		CpuFrequency:          getCpuFrequency(),
		CpuCount:              getCpuCount(),
		MemoryInfo:            getMemoryInfo(),
		DiskInfo:              getDiskInfo(),
		Disks:                 getDisks(),
		DiskInfoPerDisk:       getDiskInfoPerDisk(),
		BootTime:              getBootTime(),
		Ips:                   getIps(),
		Macs:                  getMacs(),
		NetworkConnectionType: getNetworkConnectionType(),
		NetworkSpeed:          getNetworkSpeed(),
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

func replaceIndent(str string) string {
	return strings.TrimSpace(str)
}

func strToUint(str string) uint64 {
	result, err := strconv.ParseUint(replaceIndent(str), 10, 64)

	if err != nil {
		fmt.Println("Ошибка преобразования strToUint:", err)
		return 0.0
	}

	return result
}

func strToFloat(str string) float64 {
	result, err := strconv.ParseFloat(replaceIndent(str), 64)

	if err != nil {
		fmt.Println("Ошибка преобразования strToFloat:", err)
		return 0.0
	}

	return result
}

// == todo: заменить реализацию методов на чистый GO ==

func GetCpuTemperature() float64 {
	data, err := os.ReadFile("/sys/class/thermal/thermal_zone0/temp")
	if err != nil {
		return 0.0
	}

	temp := strToFloat(string(data))

	return temp / 1000.0 // Convert to Celsius
}

func getCpuPercent() float64 {
	return strToFloat(runPythonCommand("get_cpu_percent"))
}

func getCpuPercentPerCpu() any {
	return replaceIndent(runPythonCommand("get_cpu_percent_per_cpu"))
}

func getCpuFrequency() string {
	return replaceIndent(runPythonCommand("get_cpu_frequency"))
}

func getCpuCount() uint64 {
	return strToUint(runPythonCommand("get_cpu_count"))
}

func getMemoryInfo() string {
	return replaceIndent(runPythonCommand("get_memory_info"))
}

func getDiskInfo() string {
	return replaceIndent(runPythonCommand("get_disk_info"))
}

func getDiskInfoPerDisk() any {
	return replaceIndent(runPythonCommand("get_disk_info_per_disk"))
}

func getDisks() any {
	return replaceIndent(runPythonCommand("get_disks"))
}

func getBootTime() string {
	return replaceIndent(runPythonCommand("get_boot_time"))
}

func getIps() any {
	return replaceIndent(runPythonCommand("get_ips"))
}

func getMacs() any {
	return replaceIndent(runPythonCommand("get_macs"))
}

func getNetworkConnectionType() any {
	return replaceIndent(runPythonCommand("get_network_connection_type"))
}

func getNetworkSpeed() any {
	return replaceIndent(runPythonCommand("get_network_speed"))
}
