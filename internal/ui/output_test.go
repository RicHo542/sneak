package ui

import "testing"

func TestShortPath(t *testing.T) {
	tests := []struct {
		path string
		max  int
		want string
	}{
		{path: "/short/path", max: 30, want: "/short/path"},
		{path: "/home/user/work/some-repo", max: 12, want: "...rk/some-repo"},
		{path: "global", max: 30, want: "global"},
	}
	for _, tt := range tests {
		if got := shortPath(tt.path, tt.max); got != tt.want {
			t.Errorf("shortPath(%q, %d) = %q, want %q", tt.path, tt.max, got, tt.want)
		}
	}

	got := shortPath("x", 30)
	if got != "x" {
		t.Errorf("shortPath short input = %q, want %q", got, "x")
	}
}
