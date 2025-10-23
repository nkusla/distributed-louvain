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
}

type Algorithm struct {
	MaxIterations int           `yaml:"max_iterations"`
	Timeout       time.Duration `yaml:"timeout"`
	GracePeriod   time.Duration `yaml:"grace_period"`
	DataPath      string        `yaml:"data_path"`
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

func LoadConfigFromEnv() *Config {
	return &Config{
		MachineID:     getEnv("MACHINE_ID", ""),
		Port:          getEnvInt("PORT", 8080),
		IsCoordinator: getEnvBool("IS_COORDINATOR", false),
		Coordinator:   getEnv("COORDINATOR", ""),
		Algorithm: Algorithm{
			MaxIterations: getEnvInt("MAX_ITERATIONS", 20),
			Timeout:       getEnvDuration("TIMEOUT", 60*time.Second),
			GracePeriod:   getEnvDuration("GRACE_PERIOD", 2*time.Second),
			DataPath:      getEnv("DATA_PATH", "data/karate_club.csv"),
		},
		Actors: Actors{
			Partitions:  getEnvInt("PARTITIONS", 4),
			Aggregators: getEnvInt("AGGREGATORS", 2),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
