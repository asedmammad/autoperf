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
	configPath := flag.String("config", "/etc/autoperf.conf", "path to configuration file")
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

	// Initialize current performance bias
	currentPerfBias := "balance-power"

	// Main monitoring loop
	for {
		select {
		case <-sigChan:
			log.Println("Shutting down monitor daemon")
			return nil
		default:
			if !cfg.Enabled {
				time.Sleep(time.Duration(60) * time.Second) // Sleep for 5 seconds when disabled
				continue
			}

			acPlugged, err := power.IsACPlugged()
			if err != nil {
				log.Printf("Error detecting power source: %v", err)
				continue
			}

			// Select the appropriate power config
			powerCfg := &cfg.Battery
			if acPlugged {
				powerCfg = &cfg.AC
			}

			if err := runMonitoringCycle(cfg, powerCfg, acPlugged, &currentPerfBias); err != nil {
				log.Printf("Error in monitoring cycle: %v", err)
			}
			time.Sleep(time.Duration(powerCfg.WaitBetweenUpdates) * time.Second)
		}
	}
}

func runMonitoringCycle(cfg *config.Config, powerCfg *config.PowerConfig, acPlugged bool, currentPerfBias *string) error {
	cpuLoad, err := monitor.GetCPULoad(powerCfg.CPUSampleInterval)
	if err != nil {
		return err
	}

	cpuTemp, err := monitor.GetCPUTemp()
	if err != nil {
		return err
	}

	// Determine appropriate energy_perf_bias value
	var perfBias string
	switch {
	case cpuLoad > powerCfg.HighLoadThreshold && acPlugged:
		perfBias = "performance"
	case cpuLoad > powerCfg.MediumLoadThreshold && acPlugged:
		perfBias = "balance-performance"
	case cpuTemp > powerCfg.HighTempThreshold:
		perfBias = "balance-power"
	case cpuLoad <= powerCfg.LowLoadThreshold:
		perfBias = "power"
	case !acPlugged:
		perfBias = "balance-power"
	default:
		perfBias = "balance-power"
	}

	if perfBias != *currentPerfBias {
		if err := power.SetEnergyPerfBias(perfBias); err != nil {
			return err
		}
		log.Printf("Status - Load: %.1f%%, Temp: %.1f°C, AC: %v, PerfBias: %s (changed from %s)",
			cpuLoad, cpuTemp, acPlugged, perfBias, *currentPerfBias)
		*currentPerfBias = perfBias
	} else {
		log.Printf("Status - Load: %.1f%%, Temp: %.1f°C, AC: %v, PerfBias: %s (unchanged)",
			cpuLoad, cpuTemp, acPlugged, perfBias)
	}

	return nil
}
