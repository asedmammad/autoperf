# autoperf

autoperf is a lightweight daemon that automatically adjusts CPU performance bias based on system load, temperature, and power status. It helps optimize the balance between performance and power consumption on Linux systems.

## Features

- Dynamic CPU performance bias adjustment
- Monitoring of:
  - Pressure Stall Information (`/proc/pressure/cpu`)
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
sudo cp autoperf.conf /etc/autoperf.conf
 ```

3. Install systemd service:
```bash
sudo cp systemd/autoperf.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable autoperf
sudo systemctl start autoperf
 ```

## Configuration
The configuration file is located at /etc/autoperf.conf.

Here's the default configuration:

```ini
[General]
Enabled=true
SysfsPowerPath="/sys/class/power_supply/AC*/online"
LogFile="/var/log/autoperf.log"
Metrics="PSI" # PSI or Load
Verbose=true

[Battery]
WaitBetweenUpdates=3000
CPULoadSampleInterval=10
PSILowThreshold=4.0
PSIMediumThreshold=7.0
PSIHighThreshold=10.0
LoadLowThreshold=30
LoadMediumThreshold=50
LoadHighThreshold=80
HighTempThreshold=75

[AC]
WaitBetweenUpdates=5
CPULoadSampleInterval=10
PSILowThreshold=4.0
PSIMediumThreshold=7.0
PSIHighThreshold=10.0
LoadLowThreshold=25
LoadMediumThreshold=50
LoadHighThreshold=80
HighTempThreshold=90
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
