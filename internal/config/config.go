package config

import (
	"fmt"

	"gopkg.in/ini.v1"
)

type Config struct {
	// General settings
	Enabled        bool   `ini:"Enabled,omitempty"`
	SysfsPowerPath string `ini:"SysfsPowerPath,omitempty"`
	LogFile        string `ini:"LogFile"`

	// Power-specific settings
	Battery PowerConfig `ini:"-"`
	AC      PowerConfig `ini:"-"`
}

type PowerConfig struct {
	WaitBetweenUpdates  int     `ini:"WaitBetweenUpdates"`
	CPUSampleInterval   int     `ini:"CPULoadSampleInterval"`
	LowLoadThreshold    float64 `ini:"LowLoadThreshold"`
	MediumLoadThreshold float64 `ini:"MediumLoadThreshold"`
	HighLoadThreshold   float64 `ini:"HighLoadThreshold"`
	HighTempThreshold   float64 `ini:"HighTempThreshold"`
}

func Load(path string) (*Config, error) {
	cfg, err := ini.Load(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := &Config{}

	// Load general settings
	if err := cfg.Section("General").MapTo(config); err != nil {
		return nil, fmt.Errorf("failed to parse General section: %w", err)
	}

	// Load Battery settings
	if err := cfg.Section("Battery").MapTo(&config.Battery); err != nil {
		return nil, fmt.Errorf("failed to parse Battery section: %w", err)
	}

	// Load AC settings
	if err := cfg.Section("AC").MapTo(&config.AC); err != nil {
		return nil, fmt.Errorf("failed to parse AC section: %w", err)
	}

	if err := validate(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

func validate(c *Config) error {
	if c.LogFile == "" {
		return fmt.Errorf("logFile must be specified")
	}

	// Validate Battery settings
	if err := validatePowerConfig(&c.Battery, "Battery"); err != nil {
		return err
	}

	// Validate AC settings
	if err := validatePowerConfig(&c.AC, "AC"); err != nil {
		return err
	}

	return nil
}

func validatePowerConfig(pc *PowerConfig, source string) error {
	if pc.HighLoadThreshold <= 0 || pc.HighLoadThreshold > 100 {
		return fmt.Errorf("%s highLoadThreshold must be between 0 and 100", source)
	}
	if pc.MediumLoadThreshold <= 0 || pc.MediumLoadThreshold > 100 {
		return fmt.Errorf("%s mediumLoadThreshold must be between 0 and 100", source)
	}
	if pc.LowLoadThreshold <= 0 || pc.LowLoadThreshold > 100 {
		return fmt.Errorf("%s lowLoadThreshold must be between 0 and 100", source)
	}
	if pc.HighTempThreshold <= 0 {
		return fmt.Errorf("%s highTempThreshold must be positive", source)
	}
	if pc.WaitBetweenUpdates <= 0 {
		return fmt.Errorf("%s waitBetweenUpdates must be positive", source)
	}
	if pc.CPUSampleInterval <= 0 {
		return fmt.Errorf("%s cpuSampleInterval must be positive", source)
	}
	return nil
}
