# autoperf

autoperf is a lightweight daemon that automatically adjusts CPU performance bias based on system load, temperature, and power status. It helps optimize the balance between performance and power consumption on Linux systems.

## Features

- Dynamic CPU performance bias adjustment
- Monitoring of:
  - CPU load
  - CPU temperature
  - AC power status
- Configurable thresholds and intervals
- Systemd service integration

## Installation

### Prerequisites

- Go 1.x
- Linux system with supported CPU
- Root privileges (for accessing system files and making performance adjustments)

### Building from Source

```bash
go build -ldflags "-s -w"
```

### System Installation

1. Copy the binary to system path:

```bash
sudo cp autoperf /usr/bin/
 ```

2. Create configuration directory and copy config file:
```bash
sudo mkdir -p /etc/autoperf
sudo cp config.yaml /etc/autoperf/
 ```

3. Install systemd service:
```bash
sudo cp systemd/autoperf.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable autoperf
sudo systemctl start autoperf
 ```

## Configuration
The configuration file is located at /etc/autoperf/config.yaml.

Here's an example configuration:

```yaml
highLoadThreshold: 80    # CPU load threshold for performance mode
mediumLoadThreshold: 50  # CPU load threshold for balanced performance
highTempThreshold: 80    # Temperature threshold (°C) for power saving
monitorInterval: 5       # Monitoring interval in seconds
cpuSampleInterval: 3     # CPU load sampling duration in seconds
logFile: /var/log/autoperf.log
 ```

## Performance Bias Modes
The daemon adjusts the CPU energy_perf_bias to one of these modes based on system conditions:

- performance : Maximum performance (high CPU load with AC power)
- balance-performance : Balanced with preference for performance
- balance-power : Balanced with preference for power saving
- power : Maximum power saving (low CPU load or battery power)

## Monitoring

The daemon monitors:

- CPU load percentage
- CPU temperature
- AC power status

Logs are written to the configured log file (default: /var/log/autoperf.log ).

## System Requirements
- Linux
- CPU with energy_perf_bias support and `xf86_energy_perf_policy` executable
- Access to thermal sensors ( /sys/class/thermal )
- Root privileges
