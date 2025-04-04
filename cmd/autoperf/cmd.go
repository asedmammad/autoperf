package autoperf

import (
	"flag"
	"fmt"
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

	// Initialize moving averages
	const windowSize = 10
	loadMA := make([]float64, 0, windowSize)
	tempMA := make([]float64, 0, windowSize)

	// Determine initial power config
	isAC, err := power.IsACPlugged(cfg.SysfsPowerPath)
	if err != nil {
		return fmt.Errorf("failed to check initial AC status: %v", err)
	}

	// Create ticker for regular sampling
	var ticker *time.Ticker
	var lastACStatus bool = isAC
	
	updateTicker := func() {
		if ticker != nil {
			ticker.Stop()
		}
		powerCfg := cfg.Battery
		if lastACStatus {
			powerCfg = cfg.AC
		}
		interval := time.Duration(powerCfg.WaitBetweenUpdates) * time.Millisecond
		ticker = time.NewTicker(interval)
	}
	
	updateTicker()
	defer ticker.Stop()

	// Main monitoring loop
	for {
		select {
		case <-sigChan:
			log.Println("Shutting down monitor daemon")
			return nil
		case <-ticker.C:
			if !cfg.Enabled {
				continue
			}

			if err := runReactiveMonitoring(cfg, &loadMA, &tempMA, &currentPerfBias); err != nil {
				log.Printf("Error in monitoring cycle: %v", err)
			}
			
			// Check if AC status changed
			isAC, err := power.IsACPlugged(cfg.SysfsPowerPath)
			if err != nil {
				log.Printf("Failed to check AC status: %v", err)
				continue
			}
			if isAC != lastACStatus {
				lastACStatus = isAC
				updateTicker()
			}
		}
	}
}

func runReactiveMonitoring(cfg *config.Config, loadMA, tempMA *[]float64, currentPerfBias *string) error {
    // Check AC status
    isAC, err := power.IsACPlugged(cfg.SysfsPowerPath)
    if err != nil {
        return fmt.Errorf("failed to check AC status: %v", err)
    }

    // Select appropriate power config based on AC status
    powerCfg := cfg.Battery
    if isAC {
        powerCfg = cfg.AC
    }

    // Get metrics
    var metricValue float64
    var psiStats *monitor.PSIStats

    if cfg.MetricType == "PSI" {
        psiStats, err = monitor.GetCPUPressure()
        if err != nil {
            return fmt.Errorf("failed to get PSI stats: %v", err)
        }
        metricValue = psiStats.Some.Avg10
    } else {
        cpuLoad, err := monitor.GetCPULoad(1)
        if err != nil {
            return err
        }
        metricValue = cpuLoad
    }

    // Get CPU temperature
    cpuTemp, err := monitor.GetCPUTemp()
    if err != nil {
        return err
    }

    // Update moving averages
    *tempMA = append(*tempMA, cpuTemp)
    if len(*tempMA) > 10 {
        *tempMA = (*tempMA)[1:]
    }
    avgTemp := average(*tempMA)

    // Log metrics if verbose
    if cfg.Verbose {
        if cfg.MetricType == "PSI" {
            log.Printf("Metrics - PSI: %.1f%%, Temp: %.1f°C, Power: %s",
                metricValue, avgTemp, map[bool]string{true: "AC", false: "Battery"}[isAC])
        } else {
            log.Printf("Metrics - Load: %.1f%%, Temp: %.1f°C, Power: %s",
                metricValue, avgTemp, map[bool]string{true: "AC", false: "Battery"}[isAC])
        }
    }

    // Determine appropriate energy_perf_bias value
    var perfBias string
    switch {
    case avgTemp > powerCfg.HighTempThreshold:
        perfBias = "power"
        if cfg.Verbose {
            log.Printf("Setting power bias: high temperature (%.1f°C > %.1f°C threshold)",
                avgTemp, powerCfg.HighTempThreshold)
        }
    case cfg.MetricType == "PSI" && metricValue > powerCfg.PSIHighThreshold:
        perfBias = "performance"
    case cfg.MetricType != "PSI" && metricValue > powerCfg.LoadHighThreshold:
        perfBias = "performance"
    case cfg.MetricType == "PSI" && metricValue > powerCfg.PSIMediumThreshold:
        perfBias = "balance-performance"
    case cfg.MetricType != "PSI" && metricValue > powerCfg.LoadMediumThreshold:
        perfBias = "balance-performance"
    case cfg.MetricType == "PSI" && metricValue < powerCfg.PSILowThreshold:
        perfBias = "power"
    case cfg.MetricType != "PSI" && metricValue < powerCfg.LoadLowThreshold:
        perfBias = "power"
    default:
        perfBias = "balance-power"
    }

    // Update performance bias if changed
    if perfBias != *currentPerfBias {
        if err := power.SetEnergyPerfBias(perfBias); err != nil {
            return err
        }
        log.Printf("Status - Metric: %.1f%%, Temp: %.1f°C, Power: %s, PerfBias: %s (changed from %s)",
            metricValue, avgTemp, map[bool]string{true: "AC", false: "Battery"}[isAC],
            perfBias, *currentPerfBias)
        *currentPerfBias = perfBias
    }

    return nil
}

func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}
