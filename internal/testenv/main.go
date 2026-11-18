// Package testenv points the XDG base directories to a temporary directory,
// so tests do not read or write the user's germ files.
package testenv

import (
	"os"
	"path/filepath"
	"testing"
)

// Run is meant to be called from TestMain.
func Run(m *testing.M) int {
	dir, err := os.MkdirTemp("", "germ-test-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	for env, sub := range map[string]string{
		"XDG_CONFIG_HOME": "config",
		"XDG_STATE_HOME":  "state",
		"XDG_CACHE_HOME":  "cache",
	} {
		os.Setenv(env, filepath.Join(dir, sub))
	}

	return m.Run()
}
