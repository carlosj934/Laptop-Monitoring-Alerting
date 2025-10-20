package alerts

import (
	"fmt"
	"time"

	"github.com/carlosj934/laptop-dashboard-alerting/internal/config"
	"github.com/carlosj934/laptop-dashboard-alerting/internal/metrics"
)

// AlertType reps the type of alert
type AlertType string

const (
	AlertCPU						AlertType	= "cpu"
	AlertMemoryOverall	AlertType = "memory_overall"
	AlertMemoryProcess	AlertType	= "memory_process"
	AlertDisk						AlertType	= "disk"
)

// AlertEvent reps an alert that was triggered
type AlertEvent struct {
	Timestamp				time.Time
	AlertType				AlertType
	Message					string
	CurrentValue		float64
	ThresholdValue 	float64
	ProcessName			string
	ProcessPID			int32
}

// Engine evals metrics and generates alerts
type Engine struct {
	config	*config.Config
	lastAlertTimes	map[AlertType]time.Time
	cpuBreachStart	*time.Time
}

// NewEngine creates a new alert engine
func NewEngine(cfg *config.Config) *Engine {
	return &Engine {
		config:	cfg,
		lastAlertTimes: make(map[AlertType]time.Time),
	}
}

// Evaluate checks metrics against thresholds and returns alerts
func (e *Engine) Evaluate(snapshot *metrics.MetricSnapshot) []AlertEvent {
	var alerts []AlertEvent

	//Check if we're in quiet hours
	if e.isQuietHours() {
		return alerts
	}

	// Check CPU
	if e.config.CPU.Enabled {
		if alert := e.checkCPU(snapshot); alert != nil {
			alerts = append(alerts, *alert)
		}
	}

	//Check Memory Overall
	if e.config.Memory.Enabled {
		if alert := e.checkMemoryOverall(snapshot); alert != nil {
			alerts = append(alerts, *alert)
		}
	}

	//Check Mem processes
	if e.config.Memory.Enabled {
		processAlerts := e.checkMemoryProcesses(snapshot)
		alerts = append(alerts, processAlerts...)
	}
	
	// Check Disk
	if e.config.Disk.Enabled {
		if alert := e.checkDisk(snapshot); alert != nil {
			alerts = append(alerts, *alert)
		}
	}

	return alerts
}

// checkCPU evaluates CPU threshold with sustained breach
func (e *Engine) checkCPU(snapshot *metrics.MetricSnapshot) *AlertEvent {
	threshold := e.config.CPU.ThresholdPercent

	if snapshot.CPUPercent > threshold {
		// CPU is above threshold
		if e.cpuBreachStart == nil {
			// first time breach detected
			now := snapshot.Timestamp
			e.cpuBreachStart = &now
		}

		//check if breach has been sustained long enough
		duration := snapshot.Timestamp.Sub(*e.cpuBreachStart)
		requiredDuration := time.Duration(e.config.CPU.DurationSeconds) * time.Second

		if duration >= requiredDuration {
			// sustained breach - check cooldown
			if e.canAlert(AlertCPU) {
				e.recordAlert(AlertCPU, snapshot.Timestamp)
				return &AlertEvent {
					Timestamp: snapshot.Timestamp,
					AlertType: AlertCPU,
					Message: fmt.Sprintf("CPU exceeded %.1f%% (current: %.1f%%)", threshold, snapshot.CPUPercent),
					CurrentValue: snapshot.CPUPercent,
					ThresholdValue: threshold,
				}
			}
		}
	} else {
		// CPU is below threshold - reset breach timer
		e.cpuBreachStart = nil
	}

	return nil
}

// checkMemoryOverall evaluates overall mem threshold
func (e *Engine) checkMemoryOverall(snapshot *metrics.MetricSnapshot) *AlertEvent {
	threshold := e.config.Memory.OverallThresholdPercent

	if snapshot.MemoryPercent > threshold {
		if e.canAlert(AlertMemoryOverall) {
			e.recordAlert(AlertMemoryOverall, snapshot.Timestamp)
			return &AlertEvent {
				Timestamp: snapshot.Timestamp,
				AlertType: AlertMemoryOverall,
				Message: fmt.Sprintf("Memory usage exceeded %.1f%% (current: %.1f%%)", threshold, snapshot.MemoryPercent),
				CurrentValue: snapshot.MemoryPercent,
				ThresholdValue: threshold,
			}
		}
	}

	return nil
}

// checkMemoryProcesses evaluates per-process memory tresholds
func (e *Engine) checkMemoryProcesses(snapshot *metrics.MetricSnapshot) []AlertEvent {
	var alerts []AlertEvent
	thresholdBytes := e.config.Memory.ProcessThresholdBytes
	for _, proc := range snapshot.TopProcesses {
		if proc.MemoryUsed > thresholdBytes {
			if e.canAlert(AlertMemoryProcess) {
				e.recordAlert(AlertMemoryProcess, snapshot.Timestamp)

				thresholdGB := float64(thresholdBytes) / (1024 * 1024 * 1024)
				currentGB := float64(proc.MemoryUsed) / (1024 * 1024 * 1024)
				alerts = append(alerts, AlertEvent {
					Timestamp: snapshot.Timestamp,
					AlertType: AlertMemoryProcess,
					Message: fmt.Sprintf("%s using %.1f GB (treshold %.1f GB)", proc.Name, currentGB, thresholdGB),
					CurrentValue: currentGB,
					ThresholdValue: thresholdGB,
					ProcessName: proc.Name,
					ProcessPID: proc.PID,
				})

				// only alert for one process at a time
				break
			}
		}
	}

	return alerts
}

// checkDisk evaluates disk space threshold
func (e *Engine) checkDisk(snapshot *metrics.MetricSnapshot) *AlertEvent {
	thresholdGB := e.config.Disk.MinimumFreeGB
	freeGB := float64(snapshot.DiskFree) / (1024 * 1024 * 1024)

	if freeGB < thresholdGB {
		if e.canAlert(AlertDisk) {
			e.recordAlert(AlertDisk, snapshot.Timestamp)
			return &AlertEvent {
				Timestamp: snapshot.Timestamp,
				AlertType: AlertDisk,
				Message: fmt.Sprintf("Disk C: free space below %.1f GB (current: %.1f Gb)", thresholdGB, freeGB),
				CurrentValue: freeGB,
				ThresholdValue: thresholdGB,
			}
		}
	}
	
	return nil
}

// canAlert checks if enough time has passed since last alert (cooldown)
func (e *Engine) canAlert(alertType AlertType) bool {
	lastAlert, exists := e.lastAlertTimes[alertType]
	if !exists {
		return true
	}

	cooldown := time.Duration(e.config.Cooldown.DurationMinutes) * time.Minute
	return time.Since(lastAlert) >= cooldown
}

// recordAlert records the time an alert was sent
func (e *Engine) recordAlert(alertType AlertType, timestamp time.Time) {
	e.lastAlertTimes[alertType] = timestamp
}

// isQuietHours checks if current time is within quiet hours
func (e *Engine) isQuietHours() bool {
	now := time.Now()

	start, err := parseTimeOfDay(e.config.Cooldown.QuietHoursStart)
	if err != nil {
		return false
	}

	end, err := parseTimeOfDay(e.config.Cooldown.QuietHoursEnd)
	if err != nil {
		return false
	}

	currentMinutes := now.Hour()*60 + now.Minute()

	if start < end {
		// normal range
		return currentMinutes >= start && currentMinutes < end
	} else {
		//overnight range
		return currentMinutes >= start || currentMinutes < end
	}
}

// parseTimeOfDay parses "HH:MM" format to minutes since midnight
func parseTimeOfDay(timeStr string) (int, error) {
	var hour, minute int
	_, err := fmt.Sscanf(timeStr, "%d:%d", &hour, &minute)
	if err != nil {
		return 0, err
	}
	return hour*60 + minute, nil
}
