package paths

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMigrate(t *testing.T) {
	dir := t.TempDir()
	old := filepath.Join(dir, "old.json")
	cur := filepath.Join(dir, "germ", "new.json")

	assert.NoError(t, os.WriteFile(old, []byte("old"), 0o644))

	migrate(old, cur)

	data, err := os.ReadFile(cur)
	assert.NoError(t, err)
	assert.Equal(t, "old", string(data))
	assert.NoFileExists(t, old)
}

func TestMigrateKeepsExisting(t *testing.T) {
	dir := t.TempDir()
	old := filepath.Join(dir, "old.json")
	cur := filepath.Join(dir, "new.json")

	assert.NoError(t, os.WriteFile(old, []byte("old"), 0o644))
	assert.NoError(t, os.WriteFile(cur, []byte("new"), 0o644))

	migrate(old, cur)

	data, err := os.ReadFile(cur)
	assert.NoError(t, err)
	assert.Equal(t, "new", string(data))
	assert.FileExists(t, old)
}

func TestPaths(t *testing.T) {
	for _, p := range []string{Config(), SSR(), Names(), Cache("ssm.json")} {
		assert.True(t, filepath.IsAbs(p), p)
		assert.Equal(t, "germ", filepath.Base(filepath.Dir(p)), p)
	}
}

func TestBase(t *testing.T) {
	home, err := os.UserHomeDir()
	assert.NoError(t, err)

	t.Setenv("XDG_STATE_HOME", "")
	assert.Equal(t, filepath.Join(home, ".local", "state", "germ", "profiles.json"), Names())

	t.Setenv("XDG_STATE_HOME", "relative")
	assert.Equal(t, filepath.Join(home, ".local", "state", "germ", "profiles.json"), Names())

	t.Setenv("XDG_STATE_HOME", "/tmp/state")
	assert.Equal(t, "/tmp/state/germ/profiles.json", Names())
}
