package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	HighLoadThreshold   float64 `yaml:"highLoadThreshold"`   // percentage
	MediumLoadThreshold float64 `yaml:"mediumLoadThreshold"` // percentage
	LowLoadThreshold    float64 `yaml:"lowLoadThreshold"`    // percentage
	HighTempThreshold   float64 `yaml:"highTempThreshold"`   // Celsius
	MonitorInterval     int     `yaml:"monitorInterval"`     // seconds
	CPUSampleInterval   int     `yaml:"cpuSampleInterval"`   // seconds
	LogFile             string  `yaml:"logFile"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := validate(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

func validate(c *Config) error {
	if c.HighLoadThreshold <= 0 || c.HighLoadThreshold > 100 {
		return fmt.Errorf("highLoadThreshold must be between 0 and 100")
	}
	if c.MediumLoadThreshold <= 0 || c.MediumLoadThreshold > 100 {
		return fmt.Errorf("mediumLoadThreshold must be between 0 and 100")
	}
	if c.LowLoadThreshold <= 0 || c.LowLoadThreshold > 100 {
		return fmt.Errorf("lowLoadThreshold must be between 0 and 100")
	}
	if c.HighTempThreshold <= 0 {
		return fmt.Errorf("highTempThreshold must be positive")
	}
	if c.MonitorInterval <= 0 {
		return fmt.Errorf("monitorInterval must be positive")
	}
	if c.LogFile == "" {
		return fmt.Errorf("logFile must be specified")
	}
	if c.CPUSampleInterval <= 0 {
		return fmt.Errorf("cpuSampleInterval must be positive")
	}
	return nil
}
