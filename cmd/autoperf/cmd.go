package autoperf

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/asedmammad/autoperf/internal/config"
	"github.com/asedmammad/autoperf/internal/monitor"
	"github.com/asedmammad/autoperf/internal/power"
)

func Execute() error {
	// Parse command line flags
	configPath := flag.String("config", "/etc/autoperf/config.yaml", "path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		return err
	}

	// Set up logging
	logFile, err := os.OpenFile(cfg.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Main monitoring loop
	for {
		select {
		case <-sigChan:
			log.Println("Shutting down monitor daemon")
			return nil
		default:
			if err := runMonitoringCycle(cfg); err != nil {
				log.Printf("Error in monitoring cycle: %v", err)
			}
			time.Sleep(time.Duration(cfg.MonitorInterval) * time.Second)
		}
	}
}

func runMonitoringCycle(cfg *config.Config) error {
	cpuLoad, err := monitor.GetCPULoad(cfg.CPUSampleInterval)
	if err != nil {
		return err
	}

	cpuTemp, err := monitor.GetCPUTemp()
	if err != nil {
		return err
	}

	acPlugged, err := power.IsACPlugged()
	if err != nil {
		return err
	}

	// Determine appropriate energy_perf_bias value
	var perfBias string
	switch {
	case cpuLoad > cfg.HighLoadThreshold && acPlugged:
		perfBias = "performance"
	case cpuLoad > cfg.MediumLoadThreshold && acPlugged:
		perfBias = "balance-performance"
	case cpuTemp > cfg.HighTempThreshold:
		perfBias = "balance-power"
	case cpuLoad <= cfg.LowLoadThreshold:
		perfBias = "power"
	case !acPlugged:
		perfBias = "balance-power"
	default:
		perfBias = "balance-power"
	}

	if err := power.SetEnergyPerfBias(perfBias); err != nil {
		return err
	}

	log.Printf("Status - Load: %.1f%%, Temp: %.1f°C, AC: %v, PerfBias: %s",
		cpuLoad, cpuTemp, acPlugged, perfBias)

	return nil
}
