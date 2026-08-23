package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type LocalContext struct {
	Remote      RemoteContext          `yaml:"remote"`
	Bindings    []string               `yaml:"bindings"`
	Transitions map[string]WorkflowMap `yaml:"transitions,omitempty"`
	Overwrites  ContextOverwrites      `yaml:"overwrites,omitempty"`
}

type RemoteContext struct {
	Host string `yaml:"host"`
	Type string `yaml:"type"`

	// Shared information amongst Jira and Azure
	Project string `yaml:"project"`

	// Jira Settings - Removed for now, as we are always working with
	// parent binding
	// Board string `yaml:"board,omitempty"`

	// Azure Settings - Removed for now, as we are always working wiht
	// parent binding
	// AreaPath string `yaml:"area,omitempty"`
}

type ContextOverwrites struct {
	FeatBranchName string `yaml:"branch_prefix,omitempty"`
}

// TransitionRef describes how to move a work item into a target state.
// TransitionKey is the provider-specific token required by the API
// (Jira: the numeric transition ID; Azure: the state name), while
// DisplayName is a human-readable label of the target state.
type TransitionRef struct {
	TransitionKey string `yaml:"transition_key,omitempty"`
	DisplayName   string `yaml:"display_name,omitempty"`
}

// WorkflowMap is the two-hop workflow (start, done) for a work item type.
// The map is keyed by work item type with "default" as the fallback key.
type WorkflowMap struct {
	Start TransitionRef `yaml:"start"`
	Done  TransitionRef `yaml:"done"`
}

func LoadContext(dir string) (*LocalContext, error) {
	configPath := filepath.Join(dir, LocalConfigDir, LocalConfigFile)

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: run 'sneak init' first", configPath)
	}

	var ctx LocalContext
	if err := yaml.Unmarshal(data, &ctx); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", configPath, err)
	}

	return &ctx, nil
}

func StoreLocalContext(dir string, config *LocalContext) error {
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

func (c *LocalContext) GetWorkflowByType(typeName string) (*WorkflowMap, error) {
	for configTypeName := range c.Transitions {
		if configTypeName == typeName {
			workflow := c.Transitions[configTypeName]
			return &workflow, nil
		}
	}
	return nil, fmt.Errorf("no workflow transition found for type '%s'", typeName)
}

func (c *LocalContext) GetDefaultWorkflow() (*WorkflowMap, error) {
	defaultWorkflow, ok := c.Transitions["default"]
	if !ok {
		return nil, fmt.Errorf("cannot find default workflow")
	}

	return &defaultWorkflow, nil
}

func (c *LocalContext) SetWorkflowForTaskType(dir string, typeName string, workflow WorkflowMap) {
	if c.Transitions == nil {
		c.Transitions = make(map[string]WorkflowMap)
	}
	c.Transitions[typeName] = workflow

	if err := StoreLocalContext(dir, c); err != nil {
		fmt.Printf("Warning: Failed to save local context in '%s'\n", dir)
	}
}
