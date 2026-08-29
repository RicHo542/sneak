package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type State struct {
	ProjectDir  string       `json:"project_dir"`
	Cache       Cache        `json:"cache"`
	ActiveTasks []ActiveTask `json:"active_tasks"`
}

type ActiveTask struct {
	Key         string    `json:"key"`
	Summary     string    `json:"summary"`
	Status      string    `json:"status"`
	Managed     bool      `json:"managed"`
	Branch      string    `json:"branch,omitempty"`
	ActivatedAt time.Time `json:"activated_ts"`
}

type ActiveStates struct {
	Key string
	Dir string
}

func stateDir() (string, error) {
	base, err := SneakConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "state"), nil
}

// stateFilePath resolves the state file location for a given project
func stateFilePath(projectID string) (string, error) {
	sd, err := stateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(sd, projectID+".json"), nil
}

func LoadState(projectID string) (*State, error) {
	path, err := stateFilePath(projectID)
	if err != nil {
		return nil, err
	}

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

func SaveState(projectID string, state *State) error {
	sd, err := stateDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(sd, 0700); err != nil {
		return fmt.Errorf("failed to create %s: %w", sd, err)
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	path, err := stateFilePath(projectID)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}

	return nil
}

// DiscoverStates returns all currently tracked active states.
func DiscoverStates() ([]ActiveStates, error) {
	sd, err := stateDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(sd)
	if err != nil {
		if os.IsNotExist(err) {
			return []ActiveStates{}, nil
		}
		return nil, fmt.Errorf("failed to read state dir %s: %w", sd, err)
	}

	var states []ActiveStates
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		id := strings.TrimSuffix(entry.Name(), ".json")

		state, err := LoadState(id)
		if err != nil {
			return nil, fmt.Errorf("failed to load state '%s': %w", id, err)
		}

		if state.ProjectDir == "" {
			continue
		}

		info, err := os.Stat(state.ProjectDir)
		if err != nil || !info.IsDir() {
			fmt.Printf("Skipping state '%s': project directory '%s' does not exist\n", id, state.ProjectDir)
			continue
		}

		states = append(states, ActiveStates{
			Key: id,
			Dir: state.ProjectDir,
		})
	}

	return states, nil
}

func (s *State) activeTaskIndex() map[string]int {
	idx := make(map[string]int, len(s.ActiveTasks))
	for i, at := range s.ActiveTasks {
		idx[at.Key] = i
	}
	return idx
}

func (s *State) AddActiveTasks(
	items []*CacheItem, managed bool, branch string,
) {

	if s.ActiveTasks == nil {
		s.ActiveTasks = make([]ActiveTask, 0)
	}

	itemMap := make(map[string]int, len(s.ActiveTasks)+len(items))
	for i, at := range s.ActiveTasks {
		itemMap[at.Key] = i
	}

	now := time.Now()

	for _, item := range items {
		at := ActiveTask{
			Key:         item.Key,
			Summary:     item.Summary,
			Status:      item.Status,
			Managed:     managed,
			Branch:      branch,
			ActivatedAt: now,
		}

		if idx, exists := itemMap[item.Key]; exists {
			s.ActiveTasks[idx] = at
			continue
		}

		// Add new active task and update the index map with it in case
		// there are duplicates in the input slice.
		itemMap[item.Key] = len(s.ActiveTasks)
		s.ActiveTasks = append(s.ActiveTasks, at)
	}
}

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

func (s *State) GetActiveCacheItems() ([]*CacheItem, error) {
	cacheIdx := s.Cache.indexByKey()

	var items []*CacheItem
	for _, at := range s.ActiveTasks {
		i, ok := cacheIdx[at.Key]
		if !ok {
			return nil, fmt.Errorf("active task '%s' not found in cache", at.Key)
		}
		items = append(items, &s.Cache.Items[i])
	}
	return items, nil
}

func (s *State) GetActiveTaskByKey(key string) (*ActiveTask, error) {
	for i := range s.ActiveTasks {
		if s.ActiveTasks[i].Key == key {
			return &s.ActiveTasks[i], nil
		}
	}

	return nil, fmt.Errorf("item not found in active tasks.")
}

func (s *State) GetActiveTasksByKeyBatch(keys []string) ([]*ActiveTask, error) {
	idx := s.activeTaskIndex()

	var activeTasks []*ActiveTask
	for _, key := range keys {
		i, ok := idx[key]
		if !ok {
			return nil, fmt.Errorf("item not found in active tasks.")
		}
		activeTasks = append(activeTasks, &s.ActiveTasks[i])
	}

	return activeTasks, nil
}

func (s *State) GetNonActiveCacheItems() ([]*CacheItem, error) {
	active := s.activeTaskIndex()

	var items []*CacheItem
	for i, item := range s.Cache.Items {
		if _, isActive := active[item.Key]; !isActive {
			items = append(items, &s.Cache.Items[i])
		}
	}

	return items, nil
}

func (s *State) GetActiveTasksByCacheItems(items []*CacheItem) []*ActiveTask {
	itemIndexMap := s.activeTaskIndex()

	var focusedTasks []*ActiveTask
	for _, item := range items {
		atIndex, ok := itemIndexMap[item.Key]
		if !ok {
			continue
		}

		focusedTasks = append(focusedTasks, &s.ActiveTasks[atIndex])
	}

	return focusedTasks
}
