package monitor

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"
)

func GetCPULoad(sampleInterval int) (float64, error) {
	percentage, err := cpu.Percent(time.Duration(sampleInterval)*time.Second, false)
	if err != nil {
		return 0, err
	}
	return percentage[0], nil
}

func GetCPUTemp() (float64, error) {
	// Try gopsutil first
	temps, err := host.SensorsTemperatures()
	if err == nil && len(temps) > 0 {
		for _, temp := range temps {
			if temp.Temperature > 0 {
				return temp.Temperature, nil
			}
		}
	}

	// Fallback: Try reading directly from thermal zone
	files, err := os.ReadDir("/sys/class/thermal")
	if err != nil {
		return 0, fmt.Errorf("failed to access thermal sensors: %v", err)
	}

	validTypes := []string{"acpitz", "x86_pkg_temp", "pch_skylake"}
	for _, file := range files {
		if strings.HasPrefix(file.Name(), "thermal_zone") {
			typeBytes, err := os.ReadFile(fmt.Sprintf("/sys/class/thermal/%s/type", file.Name()))
			if err != nil {
				continue
			}

			sensorType := strings.TrimSpace(string(typeBytes))
			for _, validType := range validTypes {
				if strings.Contains(strings.ToLower(sensorType), validType) {
					tempBytes, err := os.ReadFile(fmt.Sprintf("/sys/class/thermal/%s/temp", file.Name()))
					if err != nil {
						continue
					}

					var temp float64
					if _, err := fmt.Sscanf(string(tempBytes), "%f", &temp); err == nil {
						return temp / 1000.0, nil
					}
				}
			}
		}
	}

	return 0, fmt.Errorf("no valid CPU temperature readings found")
}