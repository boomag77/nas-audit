package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Scanner  ScannerConfig  `yaml:"scanner"`
	Roots    []RootConfig   `yaml:"roots"`
	Reports  ReportsConfig  `yaml:"reports"`
}

type ServerConfig struct {
	ListenAddress string `yaml:"listen_address"`
}

type DatabaseConfig struct {
	Driver string `yaml:"driver"`
	Path   string `yaml:"path"`
}

type ScannerConfig struct {
	BatchSize      int  `yaml:"batch_size"`
	FollowSymlinks bool `yaml:"follow_symlinks"`
}

type RootConfig struct {
	Name             string   `yaml:"name"`
	Path             string   `yaml:"path"`
	ReadOnlyRequired bool     `yaml:"read_only_required"`
	Exclusions       []string `yaml:"exclusions"`
}

type ReportsConfig struct {
	Directory        string `yaml:"directory"`
	GitCommitEnabled bool   `yaml:"git_commit_enabled"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c Config) Validate() error {
	if c.Database.Driver == "" {
		return errors.New("database.driver is required")
	}
	if c.Database.Driver != "sqlite" {
		return fmt.Errorf("unsupported database.driver %q", c.Database.Driver)
	}
	if c.Database.Path == "" {
		return errors.New("database.path is required")
	}
	if c.Scanner.BatchSize <= 0 {
		return errors.New("scanner.batch_size must be greater than zero")
	}
	if len(c.Roots) == 0 {
		return errors.New("at least one root is required")
	}

	rootNames := make(map[string]struct{}, len(c.Roots))
	for i, root := range c.Roots {
		if root.Name == "" {
			return fmt.Errorf("roots[%d].name is required", i)
		}
		if root.Path == "" {
			return fmt.Errorf("roots[%d].path is required", i)
		}
		if _, ok := rootNames[root.Name]; ok {
			return fmt.Errorf("duplicate root name %q", root.Name)
		}
		rootNames[root.Name] = struct{}{}
	}

	return nil
}
