package git

import "strings"

// commitTypes is the set of recognized conventional-commit types.
var commitTypes = []string{
	"feat", "fix", "chore", "docs", "refactor", "perf",
	"test", "build", "ci", "style", "revert",
}

// NormalizeMessage trims surrounding whitespace, lowercases and collapses
// internal whitespace for deduplication purposes
func NormalizeMessage(msg string) string {
	return strings.Join(strings.Fields(strings.ToLower(msg)), " ")
}

// CommitType extracts the conventional-commit type prefix
func CommitType(msg string) string {
	lower := strings.ToLower(strings.TrimSpace(msg))
	colon := strings.IndexByte(lower, ':')
	if colon <= 0 {
		return ""
	}

	prefix := strings.TrimSpace(lower[:colon])
	for _, t := range commitTypes {
		if prefix == t {
			return t
		}
	}
	return ""
}

// StripTypePrefix removes a conventional-commit "type:" prefix from the
// message for display purposes.
func StripTypePrefix(msg string) string {
	trimmed := strings.TrimSpace(msg)
	colon := strings.IndexByte(trimmed, ':')
	if colon <= 0 {
		return trimmed
	}
	return strings.TrimSpace(trimmed[colon+1:])
}

// DedupByHash removes commits with duplicate hashes, keeping the first
// occurrence and preserving order. Used to collapse the same commit reachable
// from multiple branches when using --all.
func DedupByHash(commits []Commit) []Commit {
	seen := make(map[string]struct{}, len(commits))
	deduped := make([]Commit, 0, len(commits))
	for _, c := range commits {
		if _, ok := seen[c.Hash]; ok {
			continue
		}
		seen[c.Hash] = struct{}{}
		deduped = append(deduped, c)
	}
	return deduped
}

// CompactByMessage collapses runs of consecutive commits whose normalized
// messages match into a single entry, keeping the leading commit as the
// representative row and summing the count.
func CompactByMessage(commits []Commit) []Commit {
	if len(commits) == 0 {
		return commits
	}

	compacted := make([]Commit, 0, len(commits))
	for _, c := range commits {
		normalized := NormalizeMessage(c.Message)

		if len(compacted) > 0 && NormalizeMessage(compacted[len(compacted)-1].Message) == normalized {
			last := &compacted[len(compacted)-1]
			last.Count += c.Count
			continue
		}

		compacted = append(compacted, c)
	}

	return compacted
}
