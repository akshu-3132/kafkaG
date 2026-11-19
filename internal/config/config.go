package config

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v3"
)

// Config for the system
type Config struct {
	StorageConfig StorageConfig `yaml:"storage"`
}

type StorageConfig struct {
	DirPath        string `yaml:"dir_path"`
	IndexInterval  int64  `yaml:"index_interval"`
	MaxSegmentSize int64  `yaml:"max_segment_size"`
	MaxMessageSize int64  `yaml:"max_message_size"`
}

func DefaultConfig() *Config {
	return &Config{
		StorageConfig: StorageConfig{
			DirPath:        "./data",
			IndexInterval:  1000,    // in bytes
			MaxSegmentSize: 100,     // in bytes
			MaxMessageSize: 1000000, // in bytes
		},
	}
}

func LoadConfig(yamlFilePath string) (*Config, error) {
	config := DefaultConfig()
	data, err := os.ReadFile(yamlFilePath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read the file: %w", err)
	}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal YAML: %w", err)
	}
	return config, nil
}
