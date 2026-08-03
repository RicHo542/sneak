package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type State struct {
	Cache       Cache        `json:"cache"`
	ActiveTasks []ActiveTask `json:"active_tasks"`
}

type ActiveTask struct {
	Key     string `json:"key"`
	Summary string `json:"summary"`
	Status  string `json:"status"`
	Managed bool   `json:"managed"`
	Branch  string `json:"branch,omitempty"`
}

func stateFilePath(dir string) string {
	return filepath.Join(dir, LocalConfigDir, "state.json")
}

func LoadState(dir string) (*State, error) {
	path := stateFilePath(dir)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &State{}, nil
		}
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	return &state, nil
}

func SaveState(dir string, state *State) error {
	sneakDir := filepath.Join(dir, LocalConfigDir)
	if err := os.MkdirAll(sneakDir, 0755); err != nil {
		return fmt.Errorf("failed to create %s: %w", sneakDir, err)
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	path := stateFilePath(dir)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}

	return nil
}

func (s *State) AddActiveTasks(
	items []*CacheItem, managed bool, branch string,
) {
	for _, item := range items {
		at := ActiveTask{
			Key:     item.Key,
			Summary: item.Summary,
			Status:  item.Status,
			Managed: managed,
			Branch:  branch,
		}

		s.ActiveTasks = append(s.ActiveTasks, at)
	}
}
