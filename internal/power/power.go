package power

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func IsACPlugged(sysfsPowerPath string) (bool, error) {
	matches, err := filepath.Glob(sysfsPowerPath)
	if err != nil {
		return false, fmt.Errorf("failed to glob power path: %v", err)
	}
	if len(matches) == 0 {
		return false, fmt.Errorf("no power supply found matching pattern: %s", sysfsPowerPath)
	}
	
	content, err := os.ReadFile(matches[0])
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(content)) == "1", nil
}

func SetEnergyPerfBias(value string) error {
	cmd := exec.Command("x86_energy_perf_policy", value)
	return cmd.Run()
}