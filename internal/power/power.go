package power

import (
	"os"
	"os/exec"
	"strings"
)

func IsACPlugged() (bool, error) {
	content, err := os.ReadFile("/sys/class/power_supply/AC/online")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(content)) == "1", nil
}

func SetEnergyPerfBias(value string) error {
	cmd := exec.Command("x86_energy_perf_policy", value)
	return cmd.Run()
}