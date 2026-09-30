package todos

import (
	"testing"
	"time"
)

func TestAddNotes(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	store := NewTodoStore()
	item := &Todo{Key: "sn:1", Status: TodoStatusOpen}
	store.Items["global"] = []*Todo{item}

	if err := store.AddNotes([]*Todo{item}, "   "); err != nil {
		t.Fatalf("AddNotes(empty) error: %v", err)
	}
	if len(item.Comments) != 0 {
		t.Fatalf("AddNotes(empty) appended %d notes, want 0", len(item.Comments))
	}

	if err := store.AddNotes([]*Todo{item}, "  reasoning here  "); err != nil {
		t.Fatalf("AddNotes error: %v", err)
	}
	if len(item.Comments) != 1 {
		t.Fatalf("AddNotes appended %d notes, want 1", len(item.Comments))
	}
	note := item.Comments[0]
	if note.Text != "reasoning here" {
		t.Errorf("note text = %q, want trimmed %q", note.Text, "reasoning here")
	}
	if note.At == nil || time.Since(*note.At) > time.Minute {
		t.Errorf("note timestamp missing or stale: %v", note.At)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	persisted := loaded.Items["global"][0]
	if len(persisted.Comments) != 1 {
		t.Fatalf("reloaded notes = %d, want 1", len(persisted.Comments))
	}
	if persisted.Comments[0].Text != "reasoning here" {
		t.Errorf("reloaded note text = %q, want %q", persisted.Comments[0].Text, "reasoning here")
	}

	if err := store.AddNotes(nil, "text"); err != nil {
		t.Fatalf("AddNotes(nil) error: %v", err)
	}
	if len(item.Comments) != 1 {
		t.Fatalf("AddNotes(nil) mutated notes, want still 1")
	}
}

func TestGetAllByStatus(t *testing.T) {
	store := NewTodoStore()
	store.Items["global"] = []*Todo{
		{Key: "sn:1", Status: TodoStatusOpen},
		{Key: "sn:2", Status: TodoStatusActive},
		{Key: "sn:3", Status: TodoStatusClosed},
	}
	store.Items["proj-a"] = []*Todo{
		{Key: "sn:4", Status: TodoStatusOpen},
		{Key: "sn:5", Status: TodoStatusClosed},
	}

	open := store.GetAllByStatus([]string{TodoStatusOpen})
	if len(open) != 2 {
		t.Fatalf("open buckets = %d, want 2", len(open))
	}
	byScope := make(map[string]int)
	for _, b := range open {
		byScope[b.Scope] = len(b.Items)
	}
	if byScope["global"] != 1 || byScope["proj-a"] != 1 {
		t.Errorf("open distribution = %v, want global:1 proj-a:1", byScope)
	}

	closed := store.GetAllByStatus([]string{TodoStatusClosed})
	total := 0
	for _, b := range closed {
		total += len(b.Items)
	}
	if total != 2 {
		t.Errorf("closed total = %d, want 2", total)
	}

	empty := store.GetAllByStatus([]string{"nonexistent"})
	if len(empty) != 0 {
		t.Errorf("empty status returned %d buckets, want 0", len(empty))
	}
}

func TestCoarseStatus(t *testing.T) {
	valid := map[string]string{
		"open":        TodoStatusOpen,
		"Open":        TodoStatusOpen,
		"to do":       TodoStatusOpen,
		"backlog":     TodoStatusOpen,
		"active":      TodoStatusActive,
		"In Progress": TodoStatusActive,
		"blocked":     TodoStatusActive,
		"done":        TodoStatusClosed,
		"Resolved":    TodoStatusClosed,
		"CLOSED":      TodoStatusClosed,
		"canceled":    TodoStatusClosed,
	}
	for status, want := range valid {
		if got := CoarseStatus(status); got != want {
			t.Errorf("CoarseStatus(%q) = %q, want %q", status, got, want)
		}
	}
	if got := CoarseStatus("arbitrary-state"); got != "" {
		t.Errorf("CoarseStatus(arbitrary) = %q, want empty", got)
	}
}
