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

func GetDefaultConfig() *Config {
	return &Config{
		MonitoringInterval: 5,
		CPU: CPUAlert {
						Enabled: true,
						ThresholdPercent: 85.0,
						DurationSeconds: 30,
		},
		Memory: MemoryAlert {
						Enabled: true,
						OverallThresholdPercent: 90.0,
						ProcessThresholdBytes: 4294967296, // 4GB
		},
		Disk: DiskAlert {
						Enabled: true,
						MinimumFreeGB: 20.0,
		},
		Cooldown: CooldownConfig {
						DurationMinutes: 5,
						QuietHoursStart: "23:00",
						QuietHoursEnd: "07:00",
		},
		Logging: LoggingConfig {
						Enabled: true,
						Level: "info",
						MaxFileSizeMB: 10,
						MaxBackups: 3,
		},
	} 
}

func GetConfigPath() (string, error) {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return "", fmt.Errorf("APPDATA environment variable not set")
	}

	configDir := filepath.Join(appData, "Laptop Monitor")
	configPath := filepath.Join(configDir, "config.yaml")

	return configPath, nil
}

func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		cfg := GetDefaultConfig()
		if err := Save(cfg); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		return cfg, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

func Save(cfg *Config) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
