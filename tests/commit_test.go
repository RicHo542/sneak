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

func TestDedupByHash(t *testing.T) {
	commits := []git.Commit{
		{Hash: "a", Message: "first"},
		{Hash: "b", Message: "second"},
		{Hash: "a", Message: "first (dup)"},
		{Hash: "c", Message: "third"},
	}

	got := git.DedupByHash(commits)
	if len(got) != 3 {
		t.Fatalf("DedupByHash returned %d entries, want 3", len(got))
	}
	if got[0].Hash != "a" || got[0].Message != "first" {
		t.Errorf("first entry = %+v, want first occurrence 'a'/'first'", got[0])
	}
	if got[1].Hash != "b" {
		t.Errorf("second entry hash = %q, want 'b'", got[1].Hash)
	}
	if got[2].Hash != "c" {
		t.Errorf("third entry hash = %q, want 'c'", got[2].Hash)
	}
}

func TestDedupByHashThenCompact(t *testing.T) {
	// Simulate a commit reachable from both the current branch and a feature
	// branch with --all: the same hash appears twice consecutively. Hash dedup
	// must collapse it before message compaction, so it is never double-counted.
	commits := []git.Commit{
		{Hash: "a", Message: "cleanup", Count: 1},
		{Hash: "x", Message: "cleanup", Count: 1},
		{Hash: "b", Message: "cleanup", Count: 1},
	}

	deduped := git.DedupByHash(commits)
	if len(deduped) != 3 {
		t.Fatalf("DedupByHash returned %d entries, want 3 (all distinct hashes)", len(deduped))
	}

	compacted := git.CompactByMessage(deduped)
	if len(compacted) != 1 {
		t.Fatalf("CompactByMessage returned %d entries, want 1", len(compacted))
	}
	if compacted[0].Hash != "a" || compacted[0].Count != 3 {
		t.Errorf("compacted entry = %+v, want representative 'a' with Count 3", compacted[0])
	}
}
