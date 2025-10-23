package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	MachineID     string        `yaml:"machine_id"`
	Port          int           `yaml:"port"`
	IsCoordinator bool          `yaml:"is_coordinator"`
	Coordinator   string        `yaml:"coordinator,omitempty"`
	Algorithm     Algorithm     `yaml:"algorithm"`
	Actors        Actors        `yaml:"actors"`
	Network       Network       `yaml:"network"`
	DataPath      string        `yaml:"data_path"`
}

type Algorithm struct {
	MaxIterations int           `yaml:"max_iterations"`
	Timeout       time.Duration `yaml:"timeout"`
	GracePeriod   time.Duration `yaml:"grace_period"`
}

type Actors struct {
	Partitions  int `yaml:"partitions"`
	Aggregators int `yaml:"aggregators"`
}

type Network struct {
	Peers []Peer `yaml:"peers"`
}

type Peer struct {
	ID      string `yaml:"id"`
	Address string `yaml:"address"`
}

func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}

	if config.MachineID == "" {
		return nil, fmt.Errorf("machine_id is required")
	}

	if !config.IsCoordinator && config.Coordinator == "" {
		return nil, fmt.Errorf("coordinator address is required when not running as coordinator")
	}

	if config.IsCoordinator && config.Coordinator != "" {
		return nil, fmt.Errorf("cannot specify coordinator address when running as coordinator")
	}

	return &config, nil
}
