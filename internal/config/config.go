package config

import (
	"fmt"
	"os"
	"path/filepath"
	"gopkg.in/yaml.v3"
)

type Config struct {
	MonitoringInterval	int							`yaml:"monitoring_interval"`
	CPU									CPUAlert				`yaml:"cpu"`
	Memory							MemoryAlert			`yaml:"memory"`
	Disk 								DiskAlert				`yaml:"disk"`
	Cooldown						CooldownConfig	`yaml:"cooldown"`
	Logging							LoggingConfig		`yaml:"logging"`
}

type CPUAlert struct {
	Enabled						bool		`yaml:"enabled"`
	ThresholdPercent	float64	`yaml:"threshold_percent"`
	DurationSeconds		int			`yaml:"duration_seconds"`
}

type MemoryAlert struct {
	Enabled 								bool 		`yaml:"enabled"`
	OverallThresholdPercent	float64 `yaml:"overall_threshold_percent"`
	ProcessThresholdBytes		uint64	`yaml:"process_threshold_bytes"`
}

type DiskAlert struct {
	Enabled				bool		`yaml:"enabled"`
	MinimumFreeGB	float64	`yaml:"minimum_free_gb"`
}

type CooldownConfig struct {
	DurationMinutes	int			`yaml:"duration_minutes"`
	QuietHoursStart	string	`yaml:"quiet_hours_start"`
	QuietHoursEnd		string	`yaml:"quiet_hours_end"`
}

type LoggingConfig struct {
	Enabled				bool		`yaml:"enabled"`
	Level					string	`yaml:"level"`
	MaxFileSizeMB	int			`yaml:"max_file_size_mb"`
	MaxBackups		int			`yaml:"max_backups"`
}
