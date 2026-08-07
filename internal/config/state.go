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
	return filepath.Join(dir, LocalConfigDir, LocalStateFile)
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

	if s.ActiveTasks == nil {
		s.ActiveTasks = make([]ActiveTask, 0)
	}

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

// RemoveActiveTasks removed the supplied slice of cache items from
// the active tracking in the state. This is done based on issue key
// of the cache item.
func (s *State) RemoveActiveTasks(items []*CacheItem) {
	if s.ActiveTasks == nil {
		return
	}

	removeCandidates := make(map[string]struct{}, len(items))
	for _, it := range items {
		removeCandidates[it.Key] = struct{}{}
	}

	kept := s.ActiveTasks[:0]
	for _, at := range s.ActiveTasks {
		if _, ok := removeCandidates[at.Key]; !ok {
			kept = append(kept, at)
		}
	}

	s.ActiveTasks = kept
}

// GetActiveCacheItems resolves the active tasks back to their cached work
// items. The returned pointers reference the cache entries directly, so
// mutating them also keeps the cache in sync.
func (s *State) GetActiveCacheItems() ([]*CacheItem, error) {
	var items []*CacheItem
	for _, at := range s.ActiveTasks {
		item, err := s.Cache.GetByKey(at.Key)
		if err != nil {
			return nil, fmt.Errorf("active task '%s' not found in cache: %w", at.Key, err)
		}
		items = append(items, item)
	}
	return items, nil
}
