package config

import (
	"errors"
	"os"
	"path/filepath"
)

var (
	LocalConfigDir  = ".sneak"
	LocalConfigFile = "config.yaml"
	LocalStateFile  = "state.json"
)

var ErrNoProjectFound = errors.New(
	"cannot identify sneak context: no .sneak directory found here or in any parent directory",
)

// FindProjectDir walks upward from start, returning the nearest ancestor
// directory containing a .sneak directory. An error is raised if no valid
// .sneak directory can be found.
func FindProjectDir(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, LocalConfigDir)); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir { // Top level reached - C:/ or /
			return "", ErrNoProjectFound
		}
		dir = parent
	}
}
