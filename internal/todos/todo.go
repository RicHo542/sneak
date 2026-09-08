package todos

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

type Note struct {
	At   *time.Time `json:"at"`
	Text string     `json:"text"`
}

type TodoState struct {
	NextId int `json:"next_id"`
}

type Todo struct {
	Key         string     `json:"key"`
	Pin         bool       `json:"pin"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Labels      []string   `json:"labels"`
	Notes       []Note     `json:"notes"`
	CreatedAt   time.Time  `json:"created_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	ClosedAt    *time.Time `json:"closed_at,omitempty"`
}

type TodoStore struct {
	State TodoState          `json:"state"`
	Items map[string][]*Todo `json:"items"`
}

func NewTodoStore() *TodoStore {
	return &TodoStore{
		State: TodoState{NextId: 0},
		Items: make(map[string][]*Todo),
	}
}

func Load() (*TodoStore, error) {

	filePath, err := TodoFilePath()
	if err != nil {
		return nil, err
	}

	contents, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return NewTodoStore(), nil
		}
		return nil, fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	var store TodoStore
	if err := json.Unmarshal(contents, &store); err != nil {
		return nil, fmt.Errorf("failed reading todo store: %v", err)
	}

	if store.Items == nil {
		store.Items = map[string][]*Todo{}
	}

	return &store, nil
}

func (s *TodoStore) Save(inc bool) error {
	if inc {
		s.State.NextId = s.State.NextId + 1
	}

	targetFilePath, err := TodoFilePath()
	if err != nil {
		return err
	}

	jstate, err := json.MarshalIndent(s, "", "	")
	if err != nil {
		return fmt.Errorf("failed to marshal todo store: %w", err)
	}

	if err := os.WriteFile(targetFilePath, jstate, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %s", targetFilePath, err)
	}

	return nil
}

// GetByRef attempts to retrieve a ToDo from the list based on its
// sn:x key. Returns nil if not found.
func (s *TodoStore) GetByRef(projectId string, ref string) (*Todo, error) {
	if !HasTodoRef(ref) {
		return nil, fmt.Errorf("'%s' is not a valid sn: ref", ref)
	}

	scoped, ok := s.Items[projectId]
	if !ok {
		return nil, fmt.Errorf("project not found: %s", projectId)
	}

	for _, item := range scoped {
		if item.Key == ref {
			return item, nil
		}
	}

	return nil, fmt.Errorf("todo '%s' not found", ref)
}

func (s *TodoStore) GetByRefBatch(projectId string, refs []string) ([]*Todo, error) {

	var selection []*Todo
	var missing []string
	for _, ref := range refs {
		item, err := s.GetByRef(projectId, ref)
		if err != nil {
			missing = append(missing, ref)
			continue
		}
		selection = append(selection, item)
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("todos not found: %s", strings.Join(missing, ", "))
	}

	return selection, nil
}

func (s *TodoStore) GetByStatus(projectId string, status []string) ([]*Todo, error) {
	todos, ok := s.Items[projectId]
	if !ok {
		return nil, fmt.Errorf("no todos for project '%s' found", projectId)
	}

	set := make(map[string]struct{}, len(status))
	for _, s := range status {
		set[strings.ToLower(s)] = struct{}{}
	}

	var filtered []*Todo
	for i, item := range todos {
		if _, ok := set[strings.ToLower(item.Status)]; ok {
			filtered = append(filtered, todos[i])
		}
	}

	return filtered, nil
}

func (s *TodoStore) Create(projectId string, title string, labels []string, pin bool) (string, error) {

	if s.Items == nil {
		s.Items = map[string][]*Todo{}
	}

	newId := fmt.Sprintf("sn:%d", s.State.NextId)

	s.Items[projectId] = append(s.Items[projectId], &Todo{
		Key:         newId,
		Pin:         pin,
		Title:       title,
		Description: "",
		Status:      TodoStatusOpen,
		Labels:      labels,
		CreatedAt:   time.Now(),
	})

	if err := s.Save(true); err != nil {
		return "", err
	}

	return newId, nil
}

func (s *TodoStore) UpdateStatus(projectId string, ref string, status string) error {
	if s.Items == nil {
		s.Items = map[string][]*Todo{}
	}

	if !IsValidTodoStatus(status) {
		return fmt.Errorf("invalid status: '%s'", status)
	}

	scopedTodos, ok := s.Items[projectId]
	if !ok {
		return fmt.Errorf("project '%s' not found.", projectId)
	}

	updated := false
	for _, todo := range scopedTodos {
		if todo.Key == ref {
			todo.Status = status
			updated = true
		}
	}
	if !updated {
		return fmt.Errorf("todo '%s' not found", ref)
	}

	return s.Save(false)
}

func (s *TodoStore) Close(items []*Todo) error {
	now := time.Now()
	for _, item := range items {
		if item.Status == TodoStatusClosed {
			continue
		}
		item.Status = TodoStatusClosed
		item.ClosedAt = &now
	}
	return s.Save(false)
}

// Start marks items as active. The items must have been resolved from this
// store (they mutate in place); active items are skipped idempotently.
func (s *TodoStore) Start(items []*Todo) error {
	now := time.Now()
	for _, item := range items {
		if item.Status == TodoStatusActive {
			continue
		}
		item.Status = TodoStatusActive
		item.StartedAt = &now
	}
	return s.Save(false)
}

// AddNotes appends a timestamped note to each item's note log.
func (s *TodoStore) AddNotes(items []*Todo, text string) error {
	text = strings.TrimSpace(text)
	if text == "" || len(items) == 0 {
		return nil
	}

	now := time.Now()
	for _, item := range items {
		item.Notes = append(item.Notes, Note{At: &now, Text: text})
	}
	return s.Save(false)
}

/*
func (s *TodoStore) All() []*Todo
func (s *TodoStore) Reopen(projectId string, ref string) error
func (s *TodoStore) SetPin(projectId string, ref string, pin bool) error
*/
