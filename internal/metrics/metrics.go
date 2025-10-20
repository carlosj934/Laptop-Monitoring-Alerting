package metrics

import (
	"fmt"
	"sort"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

// MetricSnapshot represents a single point-in-time collection of metrics
type MetricSnapshot struct {
	Timestamp			time.Time
	CPUPercent		float64
	MemoryUsed		uint64
	MemoryTotal		uint64
	MemoryPercent	float64
	DiskFree			uint64
	DiskTotal			uint64
	TopProcesses	[]ProcessMetric
}

// ProcessMetric represents metrics for a single process
type ProcessMetric struct {
	PID					int32
	Name				string
	MemoryUsed	uint64
}

// Collector stores system metrics
type Collector struct{}

// NewCollector() creates a new metrics collector
func NewCollector() *Collector {
	return &Collector{}
}

func (c *Collector) Collect() (*MetricSnapshot, error){
	snapshot := &MetricSnapshot {
		Timestamp: time.Now(),
	}
	
	// Collect CPU info
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU usage: %w", err)
	}
	
	if len(cpuPercent) > 0 {
		snapshot.CPUPercent = cpuPercent[0]
	}
	
	//Collect Memory info
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory info: %w", err)
	}
	snapshot.MemoryUsed = memInfo.Used
	snapshot.MemoryTotal = memInfo.Total
	snapshot.MemoryPercent = memInfo.UsedPercent

	//Collect Disk info (C: drive since we're on Windows)
	diskInfo, err := disk.Usage("C:")
	if err != nil {
		return nil, fmt.Errorf("failed to get disk info: %w", err)
	}
	snapshot.DiskFree = diskInfo.Free
	snapshot.DiskTotal = diskInfo.Total

	// Collect top 5 memory-consuming processes
	topProcs, err := c.getTopProcesses(5)
	if err != nil {
		// Don't fail completely if process enumeration fails
		// just log and move on
		topProcs = []ProcessMetric{}
	}
	snapshot.TopProcesses = topProcs

	return snapshot, nil
}


// getTopProcesses returns the top N processes by memory usage
func (c *Collector) getTopProcesses(n int) ([]ProcessMetric, error) {
	pids, err := process.Pids()
	if err != nil {
		return nil, fmt.Errorf("failed to get process list: %w", err)
	}

	var processes []ProcessMetric

	for _, pid := range pids {
		proc, err := process.NewProcess(pid)
		if err != nil {
			continue // skip procs we can't access
		}

		name, err := proc.Name()
		if err != nil {
			continue
		}

		memInfo, err := proc.MemoryInfo()
		if err != nil {
			continue
		}

		processes = append(processes, ProcessMetric {
			PID: 				pid,
			Name: 			name,
			MemoryUsed: memInfo.RSS,
		})
	}

	// sort by mem usage (descending)
	sort.Slice(processes, func(i, j int) bool {
		return processes[i].MemoryUsed > processes[j].MemoryUsed
	})

	// Return top N
	if len(processes) > n {
		processes = processes[:n]
	}

	return processes, nil
}

// FormatBytes converts bytes to human-readable format
func FormatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
