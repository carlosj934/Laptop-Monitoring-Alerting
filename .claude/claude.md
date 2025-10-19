# Laptop Monitoring System Tray App - Design Phase

## Project Overview
Native Go binary that runs as a Windows system tray application.
Monitors laptop health and sends Windows toast notifications.
Complements Windows Task Manager - not replacing it.

## Architecture Decision: System Tray Application

**Why System Tray App:**
- Direct access to Windows notification API (critical requirement)
- Runs in user session - simple notification delivery
- Visual indicator (tray icon) that monitoring is active
- Easy configuration via right-click menu
- Single .exe file - simple deployment
- Auto-starts with Windows via Startup folder

**User Experience:**
- Launches on Windows startup
- Icon appears in system tray (runs minimized/background)
- Monitors silently until threshold breached
- Shows Windows toast notification when alert triggered
- Right-click tray icon for settings/controls

## Core Components

### 1. System Tray Icon Manager
- Icon in Windows notification area
- Right-click context menu:
- View current metrics
- Configure alert thresholds
- Pause/resume monitoring
- View alert history
- Open log file
- Exit application
- Visual states: normal/warning/critical (change icon color?)

### 2. Metrics Monitor (Background Goroutine)
- Runs continuously in background
- Configurable collection interval (default: every 5 seconds)
- Collects:
- CPU usage percentage (overall)
- Memory usage (used/total, percentage)
- Disk usage (used/total, percentage)
- Per-process metrics (top consumers)
- Process lifecycle events (optional: new processes)

### 3. Alert Engine
- Evaluates metrics against configured thresholds
- Implements alert fatigue prevention:
- Cooldown periods (don't re-alert same issue too soon)
- Sustained threshold breach (must exceed for X seconds)
- Alert prioritization (warning vs critical)
- Triggers Windows toast notifications
- Logs all alerts to file

### 4. Windows Toast Notification System
- Sends native Windows 10/11 toast notifications
- Notification content:
- Alert type (CPU/Memory/Disk)
- Current value vs threshold
- Action button (optional: "Open Task Manager")
- Notification priority levels

### 5. Configuration Manager
- Reads/writes configuration file
- Config location: `%APPDATA%\LaptopMonitor\config.yaml`
- Configurable settings:
- Alert thresholds (CPU %, Memory %, Disk %)
- Monitoring interval
- Alert cooldown duration
- Process ignore list
- Sustained breach duration
- Auto-reload config on file change (optional)

### 6. Logging System
- Application log file: `%APPDATA%\LaptopMonitor\logs\monitor.log`
- Logs:
- Application lifecycle (start/stop)
- Alerts triggered (with timestamp and values)
- Errors/warnings
- Configuration changes
- Log rotation to prevent disk fill

## Key Design Questions to Answer

### 1. Alert Thresholds - What should trigger notifications?

**CPU Alerts:**
- [X] Overall CPU > 85% for 30 seconds
- [ ] Single process CPU usage?
- [ ] CPU temperature (if accessible)?

**Memory Alerts:**
- [X] Overall Memory > 90%
- [X] Single process memory > 4 GB?
- [ ] Available memory < X GB?

**Disk Alerts:**
- [ ] Disk space < 20GB free
- [ ] High disk I/O (disk thrashing)?

**Process Alerts:**
- [ ] Specific process watchlist (alert if appears/disappears)?
- [ ] Unknown process started?
- [ ] Process memory leak detection (gradual increase)?

**Your decision:** Which alerts are most useful for your workflow?

### 2. Alert Fatigue Prevention

**Cooldown Strategy:**
- [X] Don't re-alert same condition within 5 minutes
- [ ] Escalating alerts (warning → critical after sustained breach)?
- [X] Quiet hours (no alerts during certain times): 11pm - 7am

**Sustained Breach:**
- [X] Instant alert or require threshold breach for duration?: Require 30-60 seconds sustained for CPU/Memory
- [X] Instant for disk space critical

**Your decision:** How aggressive should alerting be?

### 3. Configuration Approach

**Initial Config:**
- [ ] **Option A:** Hardcoded defaults, add config file later (simpler
MVP)
- [X] **Option B:** Config file from day 1 (more flexible)

**Config Management:**
- [ ] Manual YAML editing only
- [ ] Simple settings window (checkbox form)
- [X] Both (settings window writes to YAML)

**Your decision:** Start simple or build flexibility upfront?

### 4. Metrics Collection Detail

**CPU Metrics:**
- [X] Overall CPU percentage only
- [ ] Or also: Per-core usage?
- [ ] Or also: Top 5 CPU-consuming processes?

**Memory Metrics:**
- [X] Overall usage only
- [X] Or also: Top 5 memory-consuming processes?
- [ ] Track swap/page file usage?

**Disk Metrics:**
- [X] Free space only
- [ ] Or also: Read/write rates?
- [X] Monitor all drives or just C:?

**Your decision:** How detailed should monitoring be?

### 5. User Interface Scope

**Minimal (Recommended for MVP):**
- System tray icon only
- Right-click menu for basic controls
- Toast notifications
- Config via YAML file

**Your decision:** What's the MVP scope?

## Technology Stack

### Required Go Libraries
```go
// System tray management
"github.com/getlantern/systray"  // Cross-platform system tray

// System metrics collection
"github.com/shirou/gopsutil/v3/cpu"
"github.com/shirou/gopsutil/v3/mem"
"github.com/shirou/gopsutil/v3/disk"
"github.com/shirou/gopsutil/v3/process"

// Windows toast notifications
"gopkg.in/toast.v1"  // or research alternative

// Configuration
"gopkg.in/yaml.v3"  // YAML parsing

File Structure

%APPDATA%\LaptopMonitor\
├── config.yaml           # User configuration
└── logs\
  └── monitor.log       # Application logs

%USERPROFILE%\AppData\Roaming\Microsoft\Windows\Start
Menu\Programs\Startup\
└── LaptopMonitor.lnk    # Shortcut for auto-start

Auto-Start Implementation

- Create shortcut in Windows Startup folder
- Or: Registry key HKCU\Software\Microsoft\Windows\CurrentVersion\Run

Data Models to Define

1. MetricSnapshot

type MetricSnapshot struct {
  Timestamp      time.Time
  CPUPercent     float64          // Overall CPU usage (0-100)
  MemoryUsed     uint64           // Bytes used
  MemoryTotal    uint64           // Total bytes
  MemoryPercent  float64          // Percentage (0-100)
  DiskFree       uint64           // C: drive free bytes
  DiskTotal      uint64           // C: drive total bytes
  TopProcesses   []ProcessMetric  // Top 5 memory consumers
}

type ProcessMetric struct {
  PID        int32
  Name       string
  MemoryUsed uint64  // Bytes
}

2. AlertRule

type Config struct {
  MonitoringInterval int             `yaml:"monitoring_interval"`
  CPU                CPUAlert        `yaml:"cpu"`
  Memory             MemoryAlert     `yaml:"memory"`
  Disk               DiskAlert       `yaml:"disk"`
  Cooldown           CooldownConfig  `yaml:"cooldown"`
  Logging            LoggingConfig   `yaml:"logging"`
}

type CPUAlert struct {
  Enabled          bool    `yaml:"enabled"`
  ThresholdPercent float64 `yaml:"threshold_percent"`
  DurationSeconds  int     `yaml:"duration_seconds"`
}

type MemoryAlert struct {
  Enabled                  bool    `yaml:"enabled"`
  OverallThresholdPercent  float64 `yaml:"overall_threshold_percent"`
  ProcessThresholdBytes    uint64  `yaml:"process_threshold_bytes"`
}

type DiskAlert struct {
  Enabled       bool    `yaml:"enabled"`
  MinimumFreeGB float64 `yaml:"minimum_free_gb"`
}

type CooldownConfig struct {
  DurationMinutes int    `yaml:"duration_minutes"`
  QuietHoursStart string `yaml:"quiet_hours_start"`
  QuietHoursEnd   string `yaml:"quiet_hours_end"`
}

type LoggingConfig struct {
  Enabled       bool   `yaml:"enabled"`
  Level         string `yaml:"level"`
  MaxFileSizeMB int    `yaml:"max_file_size_mb"`
  MaxBackups    int    `yaml:"max_backups"`
}

3. AlertEvent

type AlertType string

const (
  AlertCPU            AlertType = "cpu"
  AlertMemoryOverall  AlertType = "memory_overall"
  AlertMemoryProcess  AlertType = "memory_process"
  AlertDisk           AlertType = "disk"
)

type AlertEvent struct {
  Timestamp      time.Time
  AlertType      AlertType
  Message        string        // "CPU exceeded 85% (current: 92.3%)"
  CurrentValue   float64       // 92.3
  ThresholdValue float64       // 85.0

  // Optional: for process-specific alerts
  ProcessName    string        // "chrome.exe"
  ProcessPID     int32         // 1234
  }

4. Config Schema

What goes in config.yaml?
# What structure?

monitoring_interval: 5

cpu:
enabled: true
threshold_percent: 85.0
duration_seconds: 30

memory:
enabled: true
overall_threshold_percent: 90.0
process_threshold_bytes: 4294967296  # 4GB

disk:
enabled: true
minimum_free_gb: 20.0

cooldown:
duration_minutes: 5
quiet_hours_start: "23:00"
quiet_hours_end: "07:00"

logging:
enabled: true
level: "info"
max_file_size_mb: 10
max_backups: 3

Project Structure (To Design)

laptop-monitor/
├── cmd/
│   └── monitor/
│       └── main.go           # Entry point
├── internal/
│   ├── metrics/              # Metrics collection
│   ├── alerts/               # Alert engine
│   ├── config/               # Configuration management
│   ├── tray/                 # System tray UI
│   └── notifications/        # Windows notifications
├── config.yaml.example       # Example config
└── go.mod
