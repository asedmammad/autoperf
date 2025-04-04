package monitor

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"
)

type PSIStats struct {
	Some struct {
		Avg10  float64
		Avg60  float64
		Avg300 float64
		Total  float64
	}
	Full struct {
		Avg10  float64
		Avg60  float64
		Avg300 float64
		Total  float64
	}
}

func GetCPUPressure() (*PSIStats, error) {
	file, err := os.Open("/proc/pressure/cpu")
	if err != nil {
		return nil, fmt.Errorf("failed to open CPU pressure file: %v", err)
	}
	defer file.Close()

	stats := &PSIStats{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		if len(fields) < 2 {
			continue
		}

		switch fields[0] {
		case "some":
			parseAvgFields(fields[1], &stats.Some)
		case "full":
			parseAvgFields(fields[1], &stats.Full)
		}
	}

	return stats, scanner.Err()
}

func parseAvgFields(data string, target interface{}) {
	parts := strings.Split(data, " ")
	for _, part := range parts {
		kv := strings.Split(part, "=")
		if len(kv) != 2 {
			continue
		}

		value, err := strconv.ParseFloat(kv[1], 64)
		if err != nil {
			continue
		}

		switch kv[0] {
		case "avg10":
			reflect.ValueOf(target).Elem().FieldByName("Avg10").SetFloat(value)
		case "avg60":
			reflect.ValueOf(target).Elem().FieldByName("Avg60").SetFloat(value)
		case "avg300":
			reflect.ValueOf(target).Elem().FieldByName("Avg300").SetFloat(value)
		case "total":
			reflect.ValueOf(target).Elem().FieldByName("Total").SetFloat(value)
		}
	}
}

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
