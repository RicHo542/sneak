package jira

import "testing"

func TestJiraBacklogStatusName(t *testing.T) {
	backlog := []string{
		"To Do", "to do", "TODO", "Open", "Backlog", "New",
		"Not Started", "Ready for Dev", "  To Do  ",
	}
	for _, name := range backlog {
		if !jiraBacklogStatusName(name) {
			t.Errorf("jiraBacklogStatusName(%q) = false, want true", name)
		}
	}

	notBacklog := []string{
		"In Progress", "InReview", "In Review", "Done", "Closed",
		"Ready for QA", "Blocked", "",
	}
	for _, name := range notBacklog {
		if jiraBacklogStatusName(name) {
			t.Errorf("jiraBacklogStatusName(%q) = true, want false", name)
		}
	}
}

func TestClassifyTransitions(t *testing.T) {
	// Ensure discovery records the "new"/To Do category as Open, the
	// "indeterminate"/In Progress category as Start, and the done status as
	// Done.
	transitions := []jiraTransition{
		{ID: "11", Name: "Move to To Do", To: jiraStatus{Name: "To Do", StatusCategory: jiraStatusCategory{Key: "new"}}},
		{ID: "21", Name: "Start Progress", To: jiraStatus{Name: "In Progress", StatusCategory: jiraStatusCategory{Key: "indeterminate"}}},
		{ID: "31", Name: "Resolve", To: jiraStatus{Name: "Done", StatusCategory: jiraStatusCategory{Key: "done"}}},
	}

	wm := classifyTransitions(transitions)
	if wm.Open.TransitionKey != "11" || wm.Open.DisplayName != "To Do" {
		t.Errorf("Open = %+v, want key 11 / To Do", wm.Open)
	}
	if wm.Start.TransitionKey != "21" || wm.Start.DisplayName != "In Progress" {
		t.Errorf("Start = %+v, want key 21 / In Progress", wm.Start)
	}
	if wm.Done.TransitionKey != "31" || wm.Done.DisplayName != "Done" {
		t.Errorf("Done = %+v, want key 31 / Done", wm.Done)
	}
}

func TestClassifyTransitionsCustomBacklogFallback(t *testing.T) {
	// Custom workflows may place a backlog-looking status under the
	// "indeterminate" category; the name fallback still captures it as Open.
	transitions := []jiraTransition{
		{ID: "12", Name: "Move to Backlog", To: jiraStatus{Name: "Backlog", StatusCategory: jiraStatusCategory{Key: "indeterminate"}}},
		{ID: "21", Name: "Start Progress", To: jiraStatus{Name: "In Progress", StatusCategory: jiraStatusCategory{Key: "indeterminate"}}},
	}

	wm := classifyTransitions(transitions)
	if wm.Start.TransitionKey != "12" {
		t.Errorf("Start = %+v, want first indeterminate key 12", wm.Start)
	}
	if wm.Open.TransitionKey != "12" || wm.Open.DisplayName != "Backlog" {
		t.Errorf("Open = %+v, want fallback key 12 / Backlog", wm.Open)
	}
}
