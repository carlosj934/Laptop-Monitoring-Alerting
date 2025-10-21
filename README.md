# Laptop Monitor

A lightweight Windows system tray application for monitoring laptop health and sending intelligent alerts via native Windows toast notifications.

## Overview

Laptop Monitor is a native Go application that runs silently in your Windows system tray, continuously monitoring CPU, memory, and disk usage. When resource thresholds are exceeded, it sends non-intrusive Windows toast notifications to keep you informed about your system's health.

**Why Laptop Monitor?**
- **Proactive Monitoring**: Get notified before performance issues impact your work
- **Smart Alerting**: Built-in alert fatigue prevention with cooldown periods and quiet hours
- **Native Experience**: Uses Windows toast notifications - no browser required
- **Lightweight**: Minimal resource footprint - won't slow down your system
- **Zero Configuration**: Works out of the box with sensible defaults
- **Highly Customizable**: YAML-based configuration for power users

## Features

### Resource Monitoring
- **CPU Usage**: Monitors overall CPU utilization with sustained threshold detection
- **Memory Usage**: Tracks both overall memory and per-process consumption
- **Disk Space**: Monitors available disk space across all drives
- **Top Processes**: Identifies top memory-consuming processes

### Intelligent Alerting
- **Sustained Threshold Detection**: Prevents false alarms by requiring thresholds to be exceeded for a configurable duration
- **Cooldown Periods**: Prevents alert spam with configurable cooldown between repeated alerts
- **Quiet Hours**: Automatically suppresses non-critical alerts during specified hours (default: 11pm - 7am)
- **Alert Priority Levels**: Distinguishes between warning and critical alerts

### System Tray Integration
- **Persistent Tray Icon**: Visual indicator that monitoring is active
- **Context Menu**: Quick access to:
  - Current metrics snapshot
  - Pause/resume monitoring
  - Open configuration file
  - View logs
  - Exit application

### Windows Toast Notifications
- **Native Integration**: Uses Windows 10/11 notification system
- **Actionable Alerts**: Clear information about what triggered the alert
- **Current vs Threshold Values**: Shows exactly how far over threshold you are

## Installation

### Download Pre-built Binary
1. Download `laptop-monitor.exe` from the [Releases](../../releases) page
2. Place it in a permanent location (e.g., `C:\Program Files\LaptopMonitor\`)
3. Run the executable

### Build from Source
```bash
# Clone the repository
git clone https://github.com/carlosj934/laptop-dashboard-alerting.git
cd laptop-dashboard-alerting

# Build the application
go build -o laptop-monitor.exe ./cmd/monitor

# Run the application
./laptop-monitor.exe
```

### Auto-Start on Windows Login
To have Laptop Monitor start automatically with Windows:

1. Press `Win + R` and type `shell:startup`
2. Create a shortcut to `laptop-monitor.exe` in the Startup folder
3. The application will now launch automatically on login

## Configuration

Configuration is stored in `%APPDATA%\LaptopMonitor\config.yaml`. The file is created automatically with default values on first run.

### Default Configuration
```yaml
monitoring_interval: 5  # Check metrics every 5 seconds

cpu:
  enabled: true
  threshold_percent: 85.0
  duration_seconds: 30  # Must exceed threshold for 30 seconds to trigger

memory:
  enabled: true
  overall_threshold_percent: 90.0
  process_threshold_bytes: 4294967296  # 4 GB

disk:
  enabled: true
  minimum_free_gb: 20.0

cooldown:
  duration_minutes: 5  # Wait 5 minutes before re-alerting same condition
  quiet_hours_start: "23:00"
  quiet_hours_end: "07:00"

logging:
  enabled: true
  level: "info"  # Options: debug, info, warn, error
  max_file_size_mb: 10
  max_backups: 3
```

### Customization
Edit the configuration file and the changes will take effect after restarting the application.

**Common Adjustments:**
- **More Aggressive Alerts**: Lower `threshold_percent` values or reduce `duration_seconds`
- **Less Alert Fatigue**: Increase `cooldown.duration_minutes`
- **Disable Quiet Hours**: Set `quiet_hours_start` and `quiet_hours_end` to the same value
- **Disable Specific Alerts**: Set `enabled: false` for CPU, Memory, or Disk sections

## Project Structure

```
laptop-monitor/
├── cmd/
│   └── monitor/
│       └── main.go              # Application entry point
├── internal/
│   ├── alerts/
│   │   ├── engine.go           # Alert evaluation and cooldown logic
│   │   └── engine_test.go
│   ├── config/
│   │   ├── config.go           # Configuration management
│   │   └── config_test.go
│   ├── metrics/
│   │   ├── metrics.go          # System metrics collection
│   │   └── metrics_test.go
│   ├── notifications/
│   │   └── notifier.go         # Windows toast notifications
│   └── tray/
│       └── tray.go             # System tray UI and orchestration
├── go.mod
└── README.md
```

## Architecture

### Components

1. **System Tray Manager** (`internal/tray`)
   - Manages system tray icon and context menu
   - Orchestrates all components
   - Handles application lifecycle

2. **Metrics Collector** (`internal/metrics`)
   - Collects CPU, memory, and disk metrics using `gopsutil`
   - Runs in background goroutine at configurable intervals
   - Identifies top resource-consuming processes

3. **Alert Engine** (`internal/alerts`)
   - Evaluates metrics against configured thresholds
   - Implements sustained threshold detection
   - Manages cooldown periods and quiet hours
   - Prevents duplicate alerts

4. **Notification System** (`internal/notifications`)
   - Sends Windows toast notifications using `gopkg.in/toast.v1`
   - Formats alert messages with context

5. **Configuration Manager** (`internal/config`)
   - Loads and saves YAML configuration
   - Provides sensible defaults
   - Validates configuration values

### Alert Flow
```
Metrics Collection → Alert Evaluation → Cooldown Check → Quiet Hours Check → Send Notification
       ↓                     ↓                ↓                  ↓                   ↓
   Every 5s         Threshold breach?   Recently alerted?   Quiet hours?   Windows Toast
```

## Technology Stack

- **Language**: Go 1.25+
- **System Tray**: [getlantern/systray](https://github.com/getlantern/systray)
- **Metrics Collection**: [shirou/gopsutil](https://github.com/shirou/gopsutil)
- **Notifications**: [gopkg.in/toast.v1](https://gopkg.in/toast.v1)
- **Configuration**: [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3)

## Logs

Application logs are stored in `%APPDATA%\LaptopMonitor\logs\monitor.log`.

Logs include:
- Application start/stop events
- Triggered alerts with timestamps and values
- Configuration load/save operations
- Errors and warnings

Log rotation is automatic based on the `logging.max_file_size_mb` and `logging.max_backups` settings.

## Development

### Running Tests
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/alerts
go test ./internal/config
go test ./internal/metrics
```

### Building
```bash
# Development build
go build -o laptop-monitor.exe ./cmd/monitor

# Production build (smaller binary)
go build -ldflags="-s -w" -o laptop-monitor.exe ./cmd/monitor
```

## Roadmap

Future enhancements under consideration:
- [ ] Settings GUI window for configuration (no manual YAML editing)
- [ ] Per-core CPU monitoring
- [ ] Disk I/O rate monitoring
- [ ] Network bandwidth monitoring
- [ ] Process watchlist (alert when specific processes appear/disappear)
- [ ] Memory leak detection (gradual increase over time)
- [ ] Export metrics history to CSV
- [ ] Dashboard web UI for historical metrics

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is open source and available under the [MIT License](LICENSE).

## Acknowledgments

- Built with [gopsutil](https://github.com/shirou/gopsutil) for cross-platform system metrics
- System tray functionality powered by [systray](https://github.com/getlantern/systray)
- Windows notifications via [toast](https://gopkg.in/toast.v1)

## Support

If you encounter any issues or have questions:
- Open an issue on [GitHub Issues](../../issues)
- Check existing issues for solutions

---

**Note**: This application is designed for Windows 10/11. While some components may work on other platforms, full functionality requires Windows for system tray integration and toast notifications.
