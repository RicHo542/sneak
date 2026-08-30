package azure

import "testing"

func TestClassifyAzureStates(t *testing.T) {
	states := []azureWorkItemState{
		{Name: "New", Category: "Proposed"},
		{Name: "Active", Category: "InProgress"},
		{Name: "Done", Category: "Completed"},
	}

	wm := classifyAzureStates(states)
	if wm.Open.TransitionKey != "New" || wm.Open.DisplayName != "New" {
		t.Errorf("Open = %+v, want New", wm.Open)
	}
	if wm.Start.TransitionKey != "Active" || wm.Start.DisplayName != "Active" {
		t.Errorf("Start = %+v, want Active", wm.Start)
	}
	if wm.Done.TransitionKey != "Done" || wm.Done.DisplayName != "Done" {
		t.Errorf("Done = %+v, want Done", wm.Done)
	}
}

func TestClassifyAzureStatesBugResolved(t *testing.T) {
	// Bugs resolve via "Resolved" before "Completed", so Done falls back to it.
	states := []azureWorkItemState{
		{Name: "New", Category: "Proposed"},
		{Name: "Active", Category: "InProgress"},
		{Name: "Resolved", Category: "Resolved"},
	}

	wm := classifyAzureStates(states)
	if wm.Open.TransitionKey != "New" {
		t.Errorf("Open = %+v, want New", wm.Open)
	}
	if wm.Done.TransitionKey != "Resolved" {
		t.Errorf("Done = %+v, want Resolved", wm.Done)
	}
}