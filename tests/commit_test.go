package tests

import (
	"testing"

	"github.com/richo542/sneak/internal/git"
)

func TestNormalizeMessage(t *testing.T) {
	cases := map[string]string{
		"cleanup":                  "cleanup",
		"  Cleanup  config   Logs": "cleanup config logs",
		"Feat: add thing":          "feat: add thing",
	}
	for in, want := range cases {
		if got := git.NormalizeMessage(in); got != want {
			t.Errorf("NormalizeMessage(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCommitType(t *testing.T) {
	cases := map[string]string{
		"feat: first sup command": "feat",
		"FIX: crash on startup":   "fix",
		"chore(deps): bump":       "",
		"cleanup":                 "",
		"initial statement":       "",
		"":                        "",
	}
	for in, want := range cases {
		if got := git.CommitType(in); got != want {
			t.Errorf("CommitType(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestStripTypePrefix(t *testing.T) {
	cases := map[string]string{
		"feat: first sup command": "first sup command",
		"initial statement":       "initial statement",
		"":                        "",
	}
	for in, want := range cases {
		if got := git.StripTypePrefix(in); got != want {
			t.Errorf("StripTypePrefix(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCompactByMessage(t *testing.T) {
	commits := []git.Commit{
		{Hash: "a", Message: "cleanup", Count: 1},
		{Hash: "b", Message: "Cleanup", Count: 1},
		{Hash: "c", Message: "add feature", Count: 1},
		{Hash: "d", Message: "cleanup", Count: 1},
	}

	got := git.CompactByMessage(commits)
	if len(got) != 3 {
		t.Fatalf("CompactByMessage returned %d entries, want 3", len(got))
	}
	if got[0].Hash != "a" || got[0].Count != 2 {
		t.Errorf("first entry = %+v, want representative 'a' with Count 2", got[0])
	}
	if got[1].Message != "add feature" || got[1].Count != 1 {
		t.Errorf("second entry = %+v, want 'add feature' Count 1", got[1])
	}
	if got[2].Message != "cleanup" || got[2].Count != 1 {
		t.Errorf("third entry = %+v, want 'cleanup' Count 1 (non-adjacent kept separate)", got[2])
	}
}

func TestCompactByMessageEmpty(t *testing.T) {
	var commits []git.Commit
	if got := git.CompactByMessage(commits); got != nil {
		t.Errorf("CompactByMessage(nil) = %v, want nil", got)
	}
}
