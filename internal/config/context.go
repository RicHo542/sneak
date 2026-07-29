package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var (
	LocalConfigDir  = ".sneak"
	LocalConfigFile = "config.yaml"
)

type Context struct {
	Remote     RemoteContext     `yaml:"remote"`
	Bindings   []string          `yaml:"bindings"`
	Overwrites ContextOverwrites `yaml:"overwrites,omitempty"`
}

type RemoteContext struct {
	Host string `yaml:"host"`
	Type string `yaml:"type"`

	// Shared information amongst Jira and Azure
	Project string `yaml:"project"`

	// Jira Settings
	Board string `yaml:"board,omitempty"`

	// Azure Settings
	Team     string `yaml:"team,omitempty"`
	AreaPath string `yaml:"area,omitempty"`
}

type ContextOverwrites struct {
	FeatBranchName string             `yaml:"branch_prefix,omitempty"`
	Transitions    ContextTransitions `yaml:"transitions,omitempty"`
}

type ContextTransitions struct {
	Start string `yaml:"start,omitempty"`
	Close string `yaml:"close,omitempty"`
	Ship  string `yaml:"ship,omitempty"`
}

func LoadContext(dir string) (*Context, error) {
	configPath := filepath.Join(dir, LocalConfigDir, LocalConfigFile)

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: run 'sneak init' first", configPath)
	}

	var ctx Context
	if err := yaml.Unmarshal(data, &ctx); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", configPath, err)
	}

	return &ctx, nil
}

func StoreContextConfig(dir string, config *Context) error {
	info, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("expected directory path, found file at '%s'", dir)
	}

	sneakDir := filepath.Join(dir, ".sneak")
	if err := os.MkdirAll(sneakDir, 0755); err != nil {
		return fmt.Errorf("failed to create %s: %w", sneakDir, err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	configPath := filepath.Join(sneakDir, "config.yaml")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", configPath, err)
	}

	return nil
}
